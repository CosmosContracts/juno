package v31

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/CosmosContracts/juno/v31/app/upgrades"
)

const UpgradeName = "v31"

// Upgrade ships the public CosmWasm security release (wasmd v0.61.15,
// wasmvm v3.0.8 / Wasmer 7.4.2) and the ante fee-grant fixes. No module
// stores are added, deleted or renamed; the table is kept explicit so a
// future edit is a deliberate change.
var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV31UpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added:   []string{},
		Renamed: []storetypes.StoreRename{},
		Deleted: []string{},
	},
}
