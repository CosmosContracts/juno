package univ14part2

import (
	"github.com/CosmosContracts/juno/v14/app/upgrades"
	store "github.com/cosmos/cosmos-sdk/store/types"
)

// Temp
const UpgradeName = "v14_2"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV14Part2UniUpgradeHandler,
	StoreUpgrades:        store.StoreUpgrades{},
}
