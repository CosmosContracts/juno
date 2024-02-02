package v18

import (
	"fmt"

	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/CosmosContracts/juno/v19/app/keepers"
	"github.com/CosmosContracts/juno/v19/app/upgrades"
	clocktypes "github.com/CosmosContracts/juno/v19/x/clock/types"
)

// UpgradeName defines the on-chain upgrade name for the upgrade.
const UpgradeName = "v1900alpha3"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: v1900Alpha3UpgradeHandler,
	StoreUpgrades:        store.StoreUpgrades{},
}

func v1900Alpha3UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	k *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		logger.Info(fmt.Sprintf("pre migrate version map: %v", vm))
		versionMap, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return nil, err
		}
		logger.Info(fmt.Sprintf("post migrate version map: %v", versionMap))

		if err := k.ClockKeeper.SetParams(ctx, clocktypes.Params{
			ContractGasLimit: 250_000,
		}); err != nil {
			return nil, err
		}

		return versionMap, nil
	}
}
