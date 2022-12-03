package v12

import (
	tokenfactorytypes "github.com/CosmWasm/token-factory/x/tokenfactory/types"
	"github.com/CosmosContracts/juno/v12/app/upgrades"
	store "github.com/cosmos/cosmos-sdk/store/types"
	ibcfeetypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
)

// UpgradeName defines the on-chain upgrade name for the Juno v12 upgrade.
const UpgradeName = "v12"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV12UpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{tokenfactorytypes.ModuleName, ibcfeetypes.ModuleName},
	},
}
