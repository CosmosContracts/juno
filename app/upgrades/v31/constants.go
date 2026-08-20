package v31

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/CosmosContracts/juno/v31/app/upgrades"
)

const UpgradeName = "v31"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV31UpgradeHandler,
	StoreUpgrades:        storetypes.StoreUpgrades{},
}
