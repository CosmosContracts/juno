package v23

import (
	"fmt"

	icqtypes "github.com/cosmos/ibc-apps/modules/async-icq/v7/types"

	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/CosmosContracts/juno/v26/app/keepers"
	"github.com/CosmosContracts/juno/v26/app/upgrades"
)

// UpgradeName defines the on-chain upgrade name for the upgrade.
const UpgradeName = "v2300alpha1"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: v2300Alpha1UpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			// updated modules
			icqtypes.ModuleName,
		},
	},
}

func v2300Alpha1UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	_ *keepers.AppKeepers,
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

		return versionMap, err
	}
}
