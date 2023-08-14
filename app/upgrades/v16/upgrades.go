package v16

import (
	"fmt"

	buildertypes "github.com/skip-mev/pob/x/builder/types"

	icqtypes "github.com/cosmos/ibc-apps/modules/async-icq/v7/types"
	exported "github.com/cosmos/ibc-go/v7/modules/core/exported"

	// External modules
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/CosmosContracts/juno/v17/app/keepers"
	"github.com/CosmosContracts/juno/v17/app/upgrades"
)

func CreateV16UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	keepers *keepers.AppKeepers,
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

		// https://github.com/cosmos/ibc-go/blob/v7.1.0/docs/migrations/v7-to-v7_1.md
		// explicitly update the IBC 02-client params, adding the localhost client type
		params := keepers.IBCKeeper.ClientKeeper.GetParams(ctx)
		params.AllowedClients = append(params.AllowedClients, exported.Localhost)
		keepers.IBCKeeper.ClientKeeper.SetParams(ctx, params)

		// Interchain Queries
		icqParams := icqtypes.NewParams(true, nil)
		keepers.ICQKeeper.SetParams(ctx, icqParams)

		// update gov params to use a 20% initial deposit ratio, allowing us to remote the ante handler
		govParams := keepers.GovKeeper.GetParams(ctx)
		govParams.MinInitialDepositRatio = sdk.NewDec(20).Quo(sdk.NewDec(100)).String()
		if err := keepers.GovKeeper.SetParams(ctx, govParams); err != nil {
			return nil, err
		}

		// x/Staking - set minimum commission to 0.050000000000000000
		stakingParams := keepers.StakingKeeper.GetParams(ctx)
		stakingParams.MinCommissionRate = sdk.NewDecWithPrec(5, 2)
		err = keepers.StakingKeeper.SetParams(ctx, stakingParams)
		if err != nil {
			return nil, err
		}

		// x/POB
		pobAddr := keepers.AccountKeeper.GetModuleAddress(buildertypes.ModuleName)

		builderParams := buildertypes.DefaultGenesisState().GetParams()
		builderParams.EscrowAccountAddress = pobAddr
		builderParams.MaxBundleSize = 4
		builderParams.FrontRunningProtection = false
		builderParams.MinBidIncrement.Denom = nativeDenom
		builderParams.MinBidIncrement.Amount = math.NewInt(1000000)
		builderParams.ReserveFee.Denom = nativeDenom
		builderParams.ReserveFee.Amount = math.NewInt(1000000)
		if err := keepers.BuildKeeper.SetParams(ctx, builderParams); err != nil {
			return nil, err
		}

		return versionMap, err
	}
}
