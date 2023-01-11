package v12

import (
	tokenfactorytypes "github.com/CosmWasm/token-factory/x/tokenfactory/types"
	"github.com/CosmosContracts/juno/v12/app/upgrades"
	feesharetypes "github.com/CosmosContracts/juno/v12/x/feeshare/types"
	oracletypes "github.com/CosmosContracts/juno/v12/x/oracle/types"
	store "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/gaia/v8/x/globalfee"
	ibcfeetypes "github.com/cosmos/ibc-go/v4/modules/apps/29-fee/types"
	intertxtypes "github.com/cosmos/interchain-accounts/x/inter-tx/types"
)

// UpgradeName defines the on-chain upgrade name for the Juno v12 upgrade.
const UpgradeName = "v12"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV12UpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{tokenfactorytypes.ModuleName, oracletypes.ModuleName, feesharetypes.ModuleName, globalfee.ModuleName, ibcfeetypes.ModuleName, intertxtypes.ModuleName},
	},
}
