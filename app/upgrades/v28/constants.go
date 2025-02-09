package v28

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/CosmosContracts/juno/v28/app/upgrades"
)

const UpgradeName = "v28"

const (
	mevModuleAmount  = "17343396309"
	mevModuleAccount = "juno1ma4sw9m2nvtucny6lsjhh4qywvh86zdh5dlkd4"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV28UpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Deleted: []string{
			"08-wasm",
			"builder",
		},
	},
}
