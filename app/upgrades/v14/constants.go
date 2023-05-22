package v14

import (
	"github.com/CosmosContracts/juno/v15/app/upgrades"
	"github.com/CosmosContracts/juno/v15/x/globalfee"
	ibchookstypes "github.com/CosmosContracts/juno/v15/x/ibchooks/types"
	store "github.com/cosmos/cosmos-sdk/store/types"
)

// UpgradeName defines the on-chain upgrade name for the upgrade.
const UpgradeName = "v14"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV14UpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			globalfee.ModuleName,
			ibchookstypes.StoreKey,
		},
	},
}
