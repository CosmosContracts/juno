package v23

import (
	store "cosmossdk.io/store/types"

	"github.com/CosmosContracts/juno/v27/app/upgrades"
)

// UpgradeName defines the on-chain upgrade name for the upgrade.
const UpgradeName = "v23"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV23UpgradeHandler,
	StoreUpgrades:        store.StoreUpgrades{},
}
