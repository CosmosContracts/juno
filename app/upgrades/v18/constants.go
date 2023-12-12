package v18

import (
	store "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/CosmosContracts/juno/v19/app/upgrades"
	cwhooks "github.com/CosmosContracts/juno/v19/x/cw-hooks"
	feepaytypes "github.com/CosmosContracts/juno/v19/x/feepay/types"
)

// UpgradeName defines the on-chain upgrade name for the upgrade.
const UpgradeName = "v18"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV18UpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			cwhooks.ModuleName,
			feepaytypes.ModuleName,
		},
	},
}
