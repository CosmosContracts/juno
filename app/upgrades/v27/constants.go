package v27

import (
	wasmlctypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"

	storetypes "cosmossdk.io/store/types"

	"github.com/CosmosContracts/juno/v27/app/upgrades"
)

const UpgradeName = "v27"

const (
	mevModuleAmount  = "17343396309"
	mevModuleAccount = "juno1ma4sw9m2nvtucny6lsjhh4qywvh86zdh5dlkd4"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV27UpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Deleted: []string{
			wasmlctypes.ModuleName,
			"builder",
		},
	},
}
