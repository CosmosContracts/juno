package v16

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/CosmosContracts/juno/v17/app/keepers"
	"github.com/CosmosContracts/juno/v17/app/upgrades"
	clocktypes "github.com/CosmosContracts/juno/v17/x/clock/types"
	driptypes "github.com/CosmosContracts/juno/v17/x/drip/types"
)

func CreateV17UpgradeHandler(
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

		// x/Mint
		// Double blocks per year (from 6 seconds to 3 = 2x blocks per year)
		mintParams := keepers.MintKeeper.GetParams(ctx)
		mintParams.BlocksPerYear *= 2
		if err = keepers.MintKeeper.SetParams(ctx, mintParams); err != nil {
			return nil, err
		}
		logger.Info(fmt.Sprintf("updated minted blocks per year logic to %v", mintParams))

		// x/Slashing
		// Double slashing window due to double blocks per year
		slashingParams := keepers.SlashingKeeper.GetParams(ctx)
		slashingParams.SignedBlocksWindow *= 2
		if err := keepers.SlashingKeeper.SetParams(ctx, slashingParams); err != nil {
			return nil, err
		}
		logger.Info(fmt.Sprintf("updated slashing params to %v", slashingParams))

		// x/drip
		if err := keepers.DripKeeper.SetParams(ctx, driptypes.DefaultParams()); err != nil {
			return nil, err
		}

		// x/clock
		if err := keepers.ClockKeeper.SetParams(ctx, clocktypes.DefaultParams()); err != nil {
			return nil, err
		}

		return versionMap, err
	}
}
