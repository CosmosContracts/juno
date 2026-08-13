package v31

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/types/module"
)

type migrationRunner interface {
	RunMigrations(context.Context, module.Configurator, module.VersionMap) (module.VersionMap, error)
}

func CreateV31UpgradeHandler(mm *module.Manager, cfg module.Configurator) upgradetypes.UpgradeHandler {
	return createV31UpgradeHandler(mm, cfg)
}

func createV31UpgradeHandler(mm migrationRunner, cfg module.Configurator) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		return mm.RunMigrations(ctx, cfg, vm)
	}
}
