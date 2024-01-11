package v19

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	wasmlctypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	decorators "github.com/CosmosContracts/juno/v19/app/decorators"
	"github.com/CosmosContracts/juno/v19/app/keepers"
	"github.com/CosmosContracts/juno/v19/app/upgrades"
	"github.com/cometbft/cometbft/libs/log"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

const (
	Core1MultisigVestingAccount = "juno190g5j8aszqhvtg7cprmev8xcxs6csra7xnk3n3"
	CharterCouncil              = "juno1nmezpepv3lx45mndyctz2lzqxa6d9xzd2xumkxf7a6r4nxt0y95qypm6c0"
)

func CreateV19UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	k *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		nativeDenom := upgrades.GetChainsDenomToken(ctx.ChainID())
		logger.Info(fmt.Sprintf("With native denom %s", nativeDenom))

		// Run migrations
		logger.Info(fmt.Sprintf("pre migrate version map: %v", vm))
		versionMap, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return nil, err
		}
		logger.Info(fmt.Sprintf("post migrate version map: %v", versionMap))

		// Change Rate Decorator Migration
		// Ensure all Validators have a max change rate of 5%
		maxChangeRate := sdk.MustNewDecFromStr(decorators.MaxChangeRate)
		validators := k.StakingKeeper.GetAllValidators(ctx)

		for _, validator := range validators {
			if validator.Commission.MaxChangeRate.GT(maxChangeRate) {
				validator.Commission.MaxChangeRate.Set(maxChangeRate)
				k.StakingKeeper.SetValidator(ctx, validator)
			}
		}

		// TODO: ONLY DO THIS WITH MAINNET
		migrateCore1Vesting(ctx, logger, k)

		// https://github.com/cosmos/ibc-go/blob/main/docs/docs/03-light-clients/04-wasm/03-integration.md
		params := k.IBCKeeper.ClientKeeper.GetParams(ctx)
		params.AllowedClients = append(params.AllowedClients, wasmlctypes.Wasm)
		k.IBCKeeper.ClientKeeper.SetParams(ctx, params)

		return versionMap, err
	}
}

func migrateCore1Vesting(ctx sdk.Context, logger log.Logger, k *keepers.AppKeepers) {
	core1Addr := sdk.MustAccAddressFromBech32(Core1MultisigVestingAccount)
	charter := sdk.MustAccAddressFromBech32(CharterCouncil)

	core1Acc := k.AccountKeeper.GetAccount(ctx, core1Addr)
	vestingAcc, ok := core1Acc.(*vestingtypes.PeriodicVestingAccount)
	if !ok {
		panic(fmt.Errorf("core1Acc.(*vestingtypes.PeriodicVestingAccount): %+v", core1Acc))
	}

	baseAcc := vestingAcc.BaseAccount

	// TODO: move all delegations to the counsel directly? (or do they need to ICA or Authz on our chain to make it easier?)
	redelegated := completeAllRedelegations(ctx, ctx.BlockTime(), k, baseAcc.GetAddress())
	unbonded, err := unbondAllAndFinish(ctx, ctx.BlockTime(), k, baseAcc.GetAddress())
	if err != nil {
		panic(err)
	}

	fmt.Printf("redelegated: %s\n", redelegated)
	fmt.Printf("unbonded: %s\n", unbonded)

	// Send all vesting funds to the charter (must be minted first)
	currBal := k.BankKeeper.GetBalance(ctx, baseAcc.GetAddress(), "ujuno")
	diff := vestingAcc.OriginalVesting.Sub(currBal)
	if err := k.BankKeeper.MintCoins(ctx, "mint", diff); err != nil {
		panic(err)
	}
	if err := k.BankKeeper.SendCoinsFromModuleToAccount(ctx, "mint", charter, diff); err != nil {
		panic(err)
	}

	// remove all vesting from the account
	k.AccountKeeper.SetAccount(ctx, baseAcc)

	// send any current tokens to the charter council

	// transfer all balance to the charter council
	if err := k.BankKeeper.SendCoins(
		ctx,
		baseAcc.GetAddress(),
		charter,
		sdk.NewCoins(currBal)); err != nil {
		panic(err)
	}
}

// From Prop16
func completeAllRedelegations(ctx sdk.Context, now time.Time, keepers *keepers.AppKeepers, accAddr sdk.AccAddress) error {
	for _, activeRedelegation := range keepers.StakingKeeper.GetRedelegations(ctx, accAddr, 65535) {
		redelegationSrc, _ := sdk.ValAddressFromBech32(activeRedelegation.ValidatorSrcAddress)
		redelegationDst, _ := sdk.ValAddressFromBech32(activeRedelegation.ValidatorDstAddress)

		// set all entry completionTime to now so we can complete re-delegation
		for i := range activeRedelegation.Entries {
			activeRedelegation.Entries[i].CompletionTime = now
		}

		keepers.StakingKeeper.SetRedelegation(ctx, activeRedelegation)
		_, err := keepers.StakingKeeper.CompleteRedelegation(ctx, accAddr, redelegationSrc, redelegationDst)
		if err != nil {
			return err
		}
	}

	return nil
}

func unbondAllAndFinish(ctx sdk.Context, now time.Time, keepers *keepers.AppKeepers, accAddr sdk.AccAddress) (math.Int, error) {
	unbondedAmt := math.ZeroInt()

	// Unbond all delegations from the account
	for _, delegation := range keepers.StakingKeeper.GetAllDelegatorDelegations(ctx, accAddr) {
		validatorValAddr := delegation.GetValidatorAddr()
		_, found := keepers.StakingKeeper.GetValidator(ctx, validatorValAddr)
		if !found {
			continue
		}

		_, err := keepers.StakingKeeper.Undelegate(ctx, accAddr, validatorValAddr, delegation.GetShares())
		if err != nil {
			return math.ZeroInt(), err
		}
	}

	// Take all unbonding and complete them.
	for _, unbondingDelegation := range keepers.StakingKeeper.GetAllUnbondingDelegations(ctx, accAddr) {
		validatorStringAddr := unbondingDelegation.ValidatorAddress
		validatorValAddr, _ := sdk.ValAddressFromBech32(validatorStringAddr)

		// Complete unbonding delegation
		for i := range unbondingDelegation.Entries {
			unbondingDelegation.Entries[i].CompletionTime = now
			unbondedAmt = unbondedAmt.Add(unbondingDelegation.Entries[i].Balance)
		}

		keepers.StakingKeeper.SetUnbondingDelegation(ctx, unbondingDelegation)
		_, err := keepers.StakingKeeper.CompleteUnbonding(ctx, accAddr, validatorValAddr)
		if err != nil {
			return math.ZeroInt(), err
		}
	}

	return unbondedAmt, nil
}
