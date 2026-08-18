package node_test

import (
	"testing"

	"github.com/cosmos/interchaintest/v10/testutil"
	"github.com/stretchr/testify/require"

	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"
)

func TestStateSyncSpecIsIsolatedAndPreconfiguresSnapshotProviderTopology(t *testing.T) {
	originalFullNodes := *e2esuite.DefaultSpec.NumFullNodes
	originalOverrides := e2esuite.DefaultConfig.ConfigFileOverrides

	spec := stateSyncSpec()

	require.Equal(t, 2, *spec.NumFullNodes)
	require.NotSame(t, e2esuite.DefaultSpec.NumFullNodes, spec.NumFullNodes)
	require.Equal(t, originalFullNodes, *e2esuite.DefaultSpec.NumFullNodes)
	require.Equal(t, originalOverrides, e2esuite.DefaultConfig.ConfigFileOverrides)
	require.NotNil(t, spec.ChainConfig.ConfigFileOverrides)

	appToml, ok := spec.ChainConfig.ConfigFileOverrides["config/app.toml"].(testutil.Toml)
	require.True(t, ok)
	snapshotToml, ok := appToml["state-sync"].(testutil.Toml)
	require.True(t, ok)
	require.Equal(t, stateSyncSnapshotInterval, snapshotToml["snapshot-interval"])
	require.Equal(t, "custom", appToml["pruning"])
}

func TestStateSyncNodeOverridesRequireDistinctRPCProviders(t *testing.T) {
	overrides := stateSyncNodeOverrides(20, "ABC123", []string{"provider-a", "provider-b"})
	configToml := overrides["config/config.toml"].(testutil.Toml)
	stateSyncToml := configToml["statesync"].(testutil.Toml)

	require.Equal(t, true, stateSyncToml["enable"])
	require.Equal(t, "tcp://provider-a:26657,tcp://provider-b:26657", stateSyncToml["rpc_servers"])
	require.EqualValues(t, 20, stateSyncToml["trust_height"])
	require.Equal(t, "ABC123", stateSyncToml["trust_hash"])
	require.Equal(t, "1h", stateSyncToml["trust_period"])
}

func TestParseSnapshotMetadata(t *testing.T) {
	output := []byte("height: 30 format: 1 chunks: 4\nheight: 20 format: 1 chunks: 3\n")

	snapshots, err := parseSnapshotMetadata(output)
	require.NoError(t, err)
	require.Equal(t, []snapshotMetadata{
		{Height: 30, Format: 1, Chunks: 4},
		{Height: 20, Format: 1, Chunks: 3},
	}, snapshots)
}
