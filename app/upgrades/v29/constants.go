package v29

import (
	"github.com/CosmosContracts/juno/v29/app/upgrades"
)

const UpgradeName = "v29"

const (
	expeditedMinDeposit = "5000000000"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV29UpgradeHandler,
}
