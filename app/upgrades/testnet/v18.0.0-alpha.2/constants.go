package v18

import (
	store "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/CosmosContracts/juno/v27/app/upgrades"
)

// UpgradeName defines the on-chain upgrade name for the upgrade.
const UpgradeName = "v1800alpha2"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: upgrades.CreateBlankUpgradeHandler,
	StoreUpgrades:        store.StoreUpgrades{},
}
