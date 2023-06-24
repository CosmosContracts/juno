package v16

import (
	buildertypes "github.com/skip-mev/pob/x/builder/types"
	icqtypes "github.com/strangelove-ventures/async-icq/v7/types"

	store "github.com/cosmos/cosmos-sdk/store/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/nft"

	"github.com/CosmosContracts/juno/v16/app/upgrades"
)

// UpgradeName defines the on-chain upgrade name for the upgrade.
const UpgradeName = "v16"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV16UpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			// new module
			buildertypes.ModuleName,
			icqtypes.ModuleName,

			// v47 module upgrades
			crisistypes.ModuleName,
			consensustypes.ModuleName,
			nft.ModuleName,
		},
	},
}
