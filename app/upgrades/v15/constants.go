package v14

import (
	"github.com/CosmosContracts/juno/v15/app/upgrades"
	store "github.com/cosmos/cosmos-sdk/store/types"
)

// UpgradeName defines the on-chain upgrade name for the upgrade.
const UpgradeName = "v15"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV15PatchUpgradeHandler,
	StoreUpgrades:        store.StoreUpgrades{},
}
