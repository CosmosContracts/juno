package v15

import (
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/CosmosContracts/juno/v27/app/keepers"
	tokenfactorytypes "github.com/CosmosContracts/juno/v27/x/tokenfactory/types"
)

// We now charge 2 million gas * gas price to create a denom.
const NewDenomCreationGasConsume uint64 = 2_000_000

func CreateV15PatchUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		logger.Info(fmt.Sprintf("pre migrate version map: %v", vm))
		versionMap, err := mm.RunMigrations(ctx, cfg, vm)
		logger.Info(fmt.Sprintf("post migrate version map: %v", versionMap))

		// x/TokenFactory
		// Use denom creation gas consumption instead of fee for contract developers
		updatedTf := tokenfactorytypes.Params{
			DenomCreationFee:        nil,
			DenomCreationGasConsume: NewDenomCreationGasConsume,
		}

		if err := keepers.TokenFactoryKeeper.SetParams(ctx, updatedTf); err != nil {
			return versionMap, err
		}
		logger.Info(fmt.Sprintf("updated tokenfactory params to %v", updatedTf))

		return versionMap, err
	}
}
