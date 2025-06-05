package v30

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/CosmosContracts/juno/v30/app/upgrades"
	feemarkettypes "github.com/CosmosContracts/juno/v30/x/feemarket/types"
)

const UpgradeName = "v30"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV30UpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Deleted: []string{
			"globalfee",
		},
		Added: []string{
			feemarkettypes.ModuleName,
		},
	},
}
