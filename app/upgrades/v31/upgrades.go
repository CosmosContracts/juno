package v31

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/CosmosContracts/juno/v31/app/keepers"
	clocktypes "github.com/CosmosContracts/juno/v31/x/clock/types"
	cwhookstypes "github.com/CosmosContracts/juno/v31/x/cw-hooks/types"
)

type migrationRunner interface {
	RunMigrations(context.Context, module.Configurator, module.VersionMap) (module.VersionMap, error)
}

type clockParamsStore interface {
	GetParams(context.Context) clocktypes.Params
	SetParams(context.Context, clocktypes.Params) error
}

type cwHooksParamsStore interface {
	Get(context.Context) (cwhookstypes.Params, error)
	Set(context.Context, cwhookstypes.Params) error
}

func migrateLegacyContractCaps(ctx context.Context, clockStore clockParamsStore, cwHooksStore cwHooksParamsStore) error {
	clockParams := clockStore.GetParams(ctx)
	if clockParams.MaxContracts == 0 {
		clockParams.MaxContracts = clocktypes.DefaultMaxContracts
		if err := clockStore.SetParams(ctx, clockParams); err != nil {
			return err
		}
	}
	cwHooksParams, err := cwHooksStore.Get(ctx)
	if err != nil {
		return err
	}
	if cwHooksParams.MaxContracts == 0 {
		cwHooksParams.MaxContracts = cwhookstypes.DefaultMaxContracts
		if err := cwHooksStore.Set(ctx, cwHooksParams); err != nil {
			return err
		}
	}
	return nil
}

func CreateV31UpgradeHandler(mm *module.Manager, cfg module.Configurator, appKeepers *keepers.AppKeepers) upgradetypes.UpgradeHandler {
	runMigrations := createV31UpgradeHandler(mm, cfg)
	return func(ctx context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		if err := migrateLegacyContractCaps(ctx, &appKeepers.ClockKeeper, appKeepers.CWHooksKeeper.Params); err != nil {
			return nil, err
		}
		return runMigrations(ctx, plan, vm)
	}
}

func createV31UpgradeHandler(mm migrationRunner, cfg module.Configurator) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		return mm.RunMigrations(ctx, cfg, vm)
	}
}
