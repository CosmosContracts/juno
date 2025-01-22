package v25

import (
	store "cosmossdk.io/store/types"

	"github.com/CosmosContracts/juno/v27/app/upgrades"
)

// UpgradeName defines the on-chain upgrade name for the upgrade.
const UpgradeName = "v25"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV25UpgradeHandler,
	StoreUpgrades:        store.StoreUpgrades{},
}
