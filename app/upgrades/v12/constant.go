package v12

import (
	"github.com/CosmosContracts/juno/v12/app/upgrades"
	oracletypes "github.com/CosmosContracts/juno/v12/x/oracle/types"
	store "github.com/cosmos/cosmos-sdk/store/types"
)

// UpgradeName defines the on-chain upgrade name for the Juno v12 upgrade.
const UpgradeName = "v12"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV12UpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		// Oracle, TokenFactory (todo, tokenfactorytypes.ModuleName)
		Added: []string{oracletypes.ModuleName},
	},
}
