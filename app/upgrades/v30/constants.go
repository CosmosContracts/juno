package v30

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/CosmosContracts/juno/v30/app/upgrades"
	feemarkettypes "github.com/CosmosContracts/juno/v30/x/feemarket/types"
	votingsnapshottypes "github.com/CosmosContracts/juno/v30/x/voting-snapshot/types"
)

const UpgradeName = "v30"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV30UpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Deleted: []string{
			"globalfee",
			"crisis",
			"params",
			"nft",
			"feeibc",          // ICS-29 fee middleware removed in ibc-go v10
			"interchainquery", // async-icq dropped — no /v10 maintainer support
		},
		Added: []string{
			feemarkettypes.ModuleName,
			votingsnapshottypes.ModuleName,
		},
	},
}
