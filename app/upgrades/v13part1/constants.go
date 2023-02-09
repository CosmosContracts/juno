package v13

import (
	tokenfactorytypes "github.com/CosmWasm/token-factory/x/tokenfactory/types"
	"github.com/CosmosContracts/juno/v13/app/upgrades"
	feesharetypes "github.com/CosmosContracts/juno/v13/x/feeshare/types"

	store "github.com/cosmos/cosmos-sdk/store/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts/controller/types"
	ibcfeetypes "github.com/cosmos/ibc-go/v4/modules/apps/29-fee/types"

	ibchookstypes "github.com/osmosis-labs/osmosis/x/ibc-hooks/types"
	packetforwardtypes "github.com/strangelove-ventures/packet-forward-middleware/v4/router/types"
	// intertxtypes "github.com/cosmos/interchain-accounts/x/inter-tx/types"
	// oracletypes "github.com/CosmosContracts/juno/v13/x/oracle/types"
)

// UpgradeName defines the on-chain upgrade name for the upgrade.
const UpgradeName = "v13-1"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV13UpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			packetforwardtypes.StoreKey,
			feesharetypes.ModuleName,
			tokenfactorytypes.ModuleName,
			ibchookstypes.StoreKey,

			icacontrollertypes.StoreKey,
			ibcfeetypes.ModuleName,
			// oracletypes.ModuleName,
			// intertxtypes.ModuleName,
		},
	},
}
