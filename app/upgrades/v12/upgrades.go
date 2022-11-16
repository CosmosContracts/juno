package v12

import (
	"github.com/CosmosContracts/juno/v12/app/keepers"
	oracletypes "github.com/CosmosContracts/juno/v12/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// CreateV12UpgradeHandler makes an upgrade handler for v12 of Juno
func CreateV12UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Oracle
		newOracleParams := oracletypes.DefaultParams()
		keepers.OracleKeeper.SetParams(ctx, newOracleParams)

		// TokenFactory (TODO)

		return mm.RunMigrations(ctx, cfg, vm)
	}
}
