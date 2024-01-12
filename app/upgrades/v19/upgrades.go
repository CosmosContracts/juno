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

		if nativeDenom == "ujuno" {
			// Mainnet Only
			migrateCore1MultisigVesting(ctx, logger, k)
		}

		// https://github.com/cosmos/ibc-go/blob/main/docs/docs/03-light-clients/04-wasm/03-integration.md
		params := k.IBCKeeper.ClientKeeper.GetParams(ctx)
		params.AllowedClients = append(params.AllowedClients, wasmlctypes.Wasm)
		k.IBCKeeper.ClientKeeper.SetParams(ctx, params)

		return versionMap, err
	}
}

// migrateCore1Vesting moves the funds and delegations from the PeriodicVestingAccount -> the new Council (contract address).
// - Get the Core-1 multisig vesting account
// - Instantly finish all redelegations, then unbond all tokens.
// - Send all tokens to the new council (including the previously held balance)
// - Sum all future vesting periods, then mint and send those tokens to the new council.
func migrateCore1MultisigVesting(ctx sdk.Context, logger log.Logger, k *keepers.AppKeepers) {
	Core1Addr := sdk.MustAccAddressFromBech32(Core1MultisigVestingAccount)
	CouncilAddr := sdk.MustAccAddressFromBech32(CharterCouncil)

	core1Acc := k.AccountKeeper.GetAccount(ctx, Core1Addr)

	vestingAcc, ok := core1Acc.(*vestingtypes.PeriodicVestingAccount)
	if !ok {
		panic(fmt.Errorf("core1Acc.(*vestingtypes.PeriodicVestingAccount): %+v", core1Acc))
	}

	// SEND TO THE CHARTER
	prop16Core1Multisig(ctx, k, Core1Addr, CouncilAddr)

	// MINT UNVESTED TOKENS TO THE CHARTER
	mintUnvestedToCharter(ctx, k, CouncilAddr, vestingAcc)

	// REMOVE VESTING FROM THE CORE1 MULTISIG (set it to the base account, no vesting terms)
	k.AccountKeeper.SetAccount(ctx, vestingAcc.BaseAccount)
}

func mintUnvestedToCharter(ctx sdk.Context, k *keepers.AppKeepers, CouncilAddr sdk.AccAddress, vestingAcc *vestingtypes.PeriodicVestingAccount) {
	unvested := SumPeriodVestingAccountsUnvestedTokensAmount(vestingAcc)
	fmt.Printf("Core1Addr Unvested to mint to the charter: %s\n", unvested)

	coins := sdk.NewCoins(sdk.NewCoin("ujuno", unvested))

	if err := k.BankKeeper.MintCoins(ctx, "mint", coins); err != nil {
		panic(err)
	}

	if err := k.BankKeeper.SendCoinsFromModuleToAccount(ctx, "mint", CouncilAddr, coins); err != nil {
		panic(err)
	}
}

func prop16Core1Multisig(ctx sdk.Context, k *keepers.AppKeepers, Core1Addr, CouncilAddr sdk.AccAddress) {
	redelegated, err := completeAllRedelegations(ctx, ctx.BlockTime(), k, Core1Addr)
	if err != nil {
		panic(err)
	}

	unbonded, err := unbondAllAndFinish(ctx, ctx.BlockTime(), k, Core1Addr)
	if err != nil {
		panic(err)
	}

	// TODO: claim rewards?

	fmt.Printf("Core1Addr Instant Redelegations: %s\n", redelegated)
	fmt.Printf("Core1Addr Instant Unbonding: %s\n", unbonded)

	// now send these to the council
	err = k.BankKeeper.SendCoins(ctx, Core1Addr, CouncilAddr, sdk.NewCoins(k.BankKeeper.GetBalance(ctx, Core1Addr, "ujuno")))
	if err != nil {
		panic(err)
	}
}

// From Prop16
func completeAllRedelegations(ctx sdk.Context, now time.Time, keepers *keepers.AppKeepers, accAddr sdk.AccAddress) (math.Int, error) {
	redelegatedAmt := math.ZeroInt()

	for _, activeRedelegation := range keepers.StakingKeeper.GetRedelegations(ctx, accAddr, 65535) {
		redelegationSrc, _ := sdk.ValAddressFromBech32(activeRedelegation.ValidatorSrcAddress)
		redelegationDst, _ := sdk.ValAddressFromBech32(activeRedelegation.ValidatorDstAddress)

		// set all entry completionTime to now so we can complete re-delegation
		for i := range activeRedelegation.Entries {
			activeRedelegation.Entries[i].CompletionTime = now
			redelegatedAmt = redelegatedAmt.Add(math.Int(activeRedelegation.Entries[i].SharesDst))
		}

		keepers.StakingKeeper.SetRedelegation(ctx, activeRedelegation)
		_, err := keepers.StakingKeeper.CompleteRedelegation(ctx, accAddr, redelegationSrc, redelegationDst)
		if err != nil {
			return redelegatedAmt, err
		}
	}

	return redelegatedAmt, nil
}

// From Prop16
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
