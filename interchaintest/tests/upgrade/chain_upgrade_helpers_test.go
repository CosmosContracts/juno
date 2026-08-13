package upgrade_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"
)

func TestUpgradeChainSpecsPinV30AndTwoChainTopology(t *testing.T) {
	require.Equal(t, "v31", upgradeName)

	specs := upgradeChainSpecs()
	require.Len(t, specs, 2)

	for i, spec := range specs {
		require.Equal(t, "v30.0.0", spec.Version, "chain %d must start at the exact release under test", i)
		require.Equal(t, "v30.0.0", spec.ChainConfig.Images[0].Version)
		require.Equal(t, e2esuite.JunoRepo, spec.ChainConfig.Images[0].Repository)
	}

	require.Equal(t, "juno-upgrade-1", specs[0].ChainConfig.ChainID)
	require.Equal(t, "juno-upgrade-2", specs[1].ChainConfig.ChainID)
	require.NotEqual(t, specs[0].ChainConfig.ChainID, specs[1].ChainConfig.ChainID)
}
