package v12

import (
	store "cosmossdk.io/store/types"

	"github.com/CosmosContracts/juno/v27/app/upgrades"
)

const UpgradeName = "v12"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV12UpgradeHandler,
	StoreUpgrades:        store.StoreUpgrades{},
}
