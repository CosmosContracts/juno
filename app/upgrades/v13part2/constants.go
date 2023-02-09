package v13

import (
	"github.com/CosmosContracts/juno/v13/app/upgrades"
	oracletypes "github.com/CosmosContracts/juno/v13/x/oracle/types"
	store "github.com/cosmos/cosmos-sdk/store/types"
	intertxtypes "github.com/cosmos/interchain-accounts/x/inter-tx/types"
)

// UpgradeName defines the on-chain upgrade name for the upgrade.
const UpgradeName = "v13-2"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV13_2UpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			oracletypes.ModuleName,
			intertxtypes.ModuleName,
		},
	},
}
