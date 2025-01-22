package v14

import (
	ibchookstypes "github.com/cosmos/ibc-apps/modules/ibc-hooks/v8/types"

	store "cosmossdk.io/store/types"

	"github.com/CosmosContracts/juno/v27/app/upgrades"
	"github.com/CosmosContracts/juno/v27/x/globalfee"
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
