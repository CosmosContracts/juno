package v26

import (
	store "cosmossdk.io/store/types"

	"github.com/CosmosContracts/juno/v27/app/upgrades"
)

// UpgradeName defines the on-chain upgrade name for the upgrade.
const UpgradeName = "v26"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV26UpgradeHandler,
	StoreUpgrades:        store.StoreUpgrades{},
}
