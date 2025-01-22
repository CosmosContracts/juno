package v18

import (
	store "cosmossdk.io/store/types"

	"github.com/CosmosContracts/juno/v27/app/upgrades"
)

// UpgradeName defines the on-chain upgrade name for the upgrade.
const UpgradeName = "v1800alpha4"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: upgrades.CreateBlankUpgradeHandler,
	StoreUpgrades:        store.StoreUpgrades{},
}
