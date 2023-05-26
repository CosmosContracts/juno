package v15

import (
	"github.com/CosmosContracts/juno/v16/app/upgrades"
	store "github.com/cosmos/cosmos-sdk/store/types"

	globalfeetypes "github.com/CosmosContracts/juno/v16/x/globalfee/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/nft"
	icqtypes "github.com/strangelove-ventures/async-icq/v7/types"
)

// UpgradeName defines the on-chain upgrade name for the upgrade.
const UpgradeName = "v16"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV16UpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			// new modules
			icqtypes.ModuleName,
			globalfeetypes.ModuleName,

			// v47 module upgrades
			crisistypes.ModuleName,
			consensustypes.ModuleName,
			nft.ModuleName,
		},
	},
}
