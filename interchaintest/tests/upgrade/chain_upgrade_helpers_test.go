package upgrade_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"
)

func TestUpgradeChainSpecsPinV30AndTwoChainTopology(t *testing.T) {
	require.Equal(t, "v31", upgradeName)
	const expectedBaseImage = "v30.0.0@sha256:081346b118fd327afb6f688ae6d6c6a430a8ff6260d9cd56e0db06630560c4db"

	specs := upgradeChainSpecs()
	require.Len(t, specs, 2)

	for i, spec := range specs {
		require.Equal(t, expectedBaseImage, spec.Version, "chain %d must start at the exact immutable image under test", i)
		require.Equal(t, expectedBaseImage, spec.ChainConfig.Images[0].Version)
		require.Equal(t, e2esuite.JunoRepo, spec.ChainConfig.Images[0].Repository)
	}

	require.Equal(t, "juno-upgrade-1", specs[0].ChainConfig.ChainID)
	require.Equal(t, "juno-upgrade-2", specs[1].ChainConfig.ChainID)
	require.NotEqual(t, specs[0].ChainConfig.ChainID, specs[1].ChainConfig.ChainID)
}
