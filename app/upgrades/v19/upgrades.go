package v19

import (
	"fmt"
	"time"

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
	core1Acc := k.AccountKeeper.GetAccount(ctx, sdk.MustAccAddressFromBech32(Core1MultisigVestingAccount))

	vestingAcc, ok := core1Acc.(*vestingtypes.PeriodicVestingAccount)
	if !ok {
		panic(fmt.Errorf("core1Acc.(*vestingtypes.PeriodicVestingAccount): %+v", core1Acc))
	}
	fmt.Println(vestingAcc)

	// remove 1 hour from the current block time to ensure it is set
	currTime := ctx.BlockTime().Sub(time.Time{}.Add(time.Hour))
	totalTokens := uint64(0)

	vestingAcc.EndTime = int64(currTime.Seconds())

	// TODO: remove all delegations instantly from prop16 code

	// sum all tokens
	for _, period := range vestingAcc.VestingPeriods {
		for _, coin := range period.Amount {
			if coin.Denom == "ujuno" {
				totalTokens += coin.Amount.Uint64()
			}
		}
	}

	totalTokens += vestingAcc.DelegatedVesting[0].Amount.Uint64()

	fmt.Println(totalTokens)
}
