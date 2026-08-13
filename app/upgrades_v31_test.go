package app

import (
	"testing"

	"github.com/stretchr/testify/require"

	v31 "github.com/CosmosContracts/juno/v31/app/upgrades/v31"
)

func TestV31UpgradeRegistered(t *testing.T) {
	matches := 0
	for _, upgrade := range Upgrades {
		if upgrade.UpgradeName == v31.UpgradeName {
			matches++
			require.Equal(t, v31.Upgrade.StoreUpgrades, upgrade.StoreUpgrades)
			require.NotNil(t, upgrade.CreateUpgradeHandler)
		}
	}
	require.Equal(t, 1, matches)
}
