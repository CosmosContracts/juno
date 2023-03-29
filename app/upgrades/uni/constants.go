package v14

import (
	"github.com/CosmosContracts/juno/v14/app/upgrades"
	store "github.com/cosmos/cosmos-sdk/store/types"
)

// UpgradeName is for temporary Uni upgrades for state breaking features
// before an official mainnet release, but after an initial alpha upgrade.
const UpgradeName = "uni14_2"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUniUpgradeHandler,
	StoreUpgrades:        store.StoreUpgrades{},
}
