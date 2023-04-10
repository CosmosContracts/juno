package v15

import (
	"fmt"

	"github.com/CosmosContracts/juno/v15/app/keepers"

	"github.com/CosmosContracts/juno/v15/app/upgrades"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	tokenfactorytypes "github.com/CosmosTokenFactory/token-factory/x/tokenfactory/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func CreateV15UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		nativeDenom := upgrades.GetChainsDenomToken(ctx.ChainID())
		logger.Info(fmt.Sprintf("With native denom %s", nativeDenom))

		logger.Info(fmt.Sprintf("pre migrate version map: %v", vm))

		// Run migrations
		versionMap, err := mm.RunMigrations(ctx, cfg, vm)

		logger.Info(fmt.Sprintf("post migrate version map: %v", versionMap))

		// Double blocks per year (from 6 seconds to 3 = 2x blocks per year)
		mintParams := keepers.MintKeeper.GetParams(ctx)
		mintParams.BlocksPerYear *= 2
		keepers.MintKeeper.SetParams(ctx, mintParams)
		logger.Info(fmt.Sprintf("updated minted blocks per year logic to %v", mintParams))

		// Add new TokenFactory param
		updatedTf := tokenfactorytypes.Params{
			DenomCreationGasConsume: 2_000_000,
		}
		keepers.TokenFactoryKeeper.SetParams(ctx, updatedTf)
		logger.Info(fmt.Sprintf("updated tokenfactory params to %v", updatedTf))

		return versionMap, err
	}
}
