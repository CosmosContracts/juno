package v14

import (
	"github.com/CosmosContracts/juno/v14/app/upgrades"
	store "github.com/cosmos/cosmos-sdk/store/types"
)

// UpgradeName defines the on-chain upgrade name for the upgrade.
const UpgradeName = "v14_2_0"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV14_2_0UpgradeHandler,
	StoreUpgrades:        store.StoreUpgrades{},
}
