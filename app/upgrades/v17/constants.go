package v16

import (
	store "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/CosmosContracts/juno/v17/app/upgrades"
)

// UpgradeName defines the on-chain upgrade name for the upgrade.
const UpgradeName = "v17"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV17UpgradeHandler,
	StoreUpgrades:        store.StoreUpgrades{},
}
