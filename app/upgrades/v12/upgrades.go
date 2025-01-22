package v12

import (
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/CosmosContracts/juno/v27/app/keepers"
)

func CreateV12UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	_ *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		return mm.RunMigrations(ctx, cfg, vm)
	}
}
