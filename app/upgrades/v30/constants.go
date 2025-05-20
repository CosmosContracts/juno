package v30

import (
	"github.com/CosmosContracts/juno/v30/app/upgrades"
)

const UpgradeName = "v30"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV30UpgradeHandler,
}
