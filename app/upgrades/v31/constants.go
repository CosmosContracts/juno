package v31

import (
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/CosmosContracts/juno/v31/app/keepers"
	"github.com/CosmosContracts/juno/v31/app/upgrades"
)

const UpgradeName = "v31"

var Upgrade = upgrades.Upgrade{
	UpgradeName: UpgradeName,
	CreateUpgradeHandler: func(mm *module.Manager, cfg module.Configurator, appKeepers *keepers.AppKeepers) upgradetypes.UpgradeHandler {
		return CreateV31UpgradeHandler(mm, cfg, appKeepers)
	},
	StoreUpgrades: storetypes.StoreUpgrades{},
}
