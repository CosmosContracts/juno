package v19

import (
	store "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/CosmosContracts/juno/v19/app/upgrades"
)

const (
	Core1MultisigVestingAccount = "juno190g5j8aszqhvtg7cprmev8xcxs6csra7xnk3n3"
	CharterCouncil              = "juno1nmezpepv3lx45mndyctz2lzqxa6d9xzd2xumkxf7a6r4nxt0y95qypm6c0"
)

// UpgradeName defines the on-chain upgrade name for the upgrade.
const UpgradeName = "v20"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV20UpgradeHandler,
	StoreUpgrades:        store.StoreUpgrades{},
}
