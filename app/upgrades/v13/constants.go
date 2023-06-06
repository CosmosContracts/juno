package v13

import (
	"github.com/CosmosContracts/juno/v16/app/upgrades"
	feesharetypes "github.com/CosmosContracts/juno/v16/x/feeshare/types"
	ibchookstypes "github.com/CosmosContracts/juno/v16/x/ibchooks/types"
	tokenfactorytypes "github.com/CosmosContracts/juno/v16/x/tokenfactory/types"
	store "github.com/cosmos/cosmos-sdk/store/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	ibcfeetypes "github.com/cosmos/ibc-go/v7/modules/apps/29-fee/types"
	packetforwardtypes "github.com/strangelove-ventures/packet-forward-middleware/v7/router/types"
)

// UpgradeName defines the on-chain upgrade name for the upgrade.
const UpgradeName = "v13"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV13UpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			tokenfactorytypes.ModuleName,
			feesharetypes.ModuleName,
			ibcfeetypes.ModuleName,
			ibchookstypes.StoreKey,
			packetforwardtypes.StoreKey,
			icacontrollertypes.StoreKey,
		},
	},
}
