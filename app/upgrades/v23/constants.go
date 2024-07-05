package v23

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	icqtypes "github.com/cosmos/ibc-apps/modules/async-icq/v7/types"

	"github.com/CosmosContracts/juno/v23/app/upgrades"
)

// UpgradeName defines the on-chain upgrade name for the upgrade.
const UpgradeName = "v23"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV23UpgradeHandler,
	StoreUpgrades:        store.StoreUpgrades{
		Added: []string{
			// updated modules
			icqtypes.ModuleName,
		},
	},
}
