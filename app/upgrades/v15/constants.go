package v15

import (
	"github.com/CosmosContracts/juno/v15/app/upgrades"
	store "github.com/cosmos/cosmos-sdk/store/types"

	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	icqtypes "github.com/strangelove-ventures/async-icq/v7/types"
)

// UpgradeName defines the on-chain upgrade name for the upgrade.
const UpgradeName = "v15"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV15UpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			// new module
			icqtypes.ModuleName,

			// v47 module upgrades
			crisistypes.ModuleName,
			consensustypes.ModuleName,
		},
	},
}
