package v21

import (
	wasmlctypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"

	store "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/CosmosContracts/juno/v21/app/upgrades"
)

// UpgradeName defines the on-chain upgrade name for the upgrade.
const UpgradeName = "v21"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV21UpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			wasmlctypes.ModuleName,
		},
	},
}
