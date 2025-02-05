package v17

import (
	store "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/CosmosContracts/juno/v27/app/upgrades"
	clocktypes "github.com/CosmosContracts/juno/v27/x/clock/types"
	driptypes "github.com/CosmosContracts/juno/v27/x/drip/types"
)

// UpgradeName defines the on-chain upgrade name for the upgrade.
const UpgradeName = "v17"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV17UpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			driptypes.ModuleName,
			clocktypes.ModuleName,
		},
	},
}
