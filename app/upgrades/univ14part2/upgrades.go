package univ14part2

import (
	"github.com/CosmosContracts/juno/v14/app/keepers"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// Basic migration just to ensure all our small non state breaking changes
// do not end up causing app hashing issues :) purely for safety on testnet
func CreateV14Part2UniUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	_ *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		return mm.RunMigrations(ctx, cfg, vm)
	}
}
