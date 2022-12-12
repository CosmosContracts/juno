package v12

import (
	tokenfactorytypes "github.com/CosmWasm/token-factory/x/tokenfactory/types"
	"github.com/CosmosContracts/juno/v12/app/upgrades"
	feesharetypes "github.com/CosmosContracts/juno/v12/x/feeshare/types"
	store "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/gaia/v8/x/globalfee"
)

// UpgradeName defines the on-chain upgrade name for the Juno v12 upgrade.
const UpgradeName = "v12"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV12UpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{tokenfactorytypes.ModuleName, feesharetypes.ModuleName, globalfee.ModuleName},
	},
}
