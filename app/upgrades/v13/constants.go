package v13

import (
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/types"
	ibchookstypes "github.com/cosmos/ibc-apps/modules/ibc-hooks/v8/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	ibcfeetypes "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/types"

	store "cosmossdk.io/store/types"

	"github.com/CosmosContracts/juno/v27/app/upgrades"
	feesharetypes "github.com/CosmosContracts/juno/v27/x/feeshare/types"
	tokenfactorytypes "github.com/CosmosContracts/juno/v27/x/tokenfactory/types"
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
