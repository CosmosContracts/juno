package interchaintest

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
)

const stateSyncSnapshotInterval = 10

func TestJunoStateSync(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	t.Parallel()

	configFileOverrides := make(map[string]any)
	appTomlOverrides := make(testutil.Toml)

	// state sync snapshots every stateSyncSnapshotInterval blocks.
	stateSync := make(testutil.Toml)
	stateSync["snapshot-interval"] = stateSyncSnapshotInterval
	appTomlOverrides["state-sync"] = stateSync

	// state sync snapshot interval must be a multiple of pruning keep every interval.
	appTomlOverrides["pruning"] = "custom"
	appTomlOverrides["pruning-keep-recent"] = stateSyncSnapshotInterval
	appTomlOverrides["pruning-keep-every"] = stateSyncSnapshotInterval
	appTomlOverrides["pruning-interval"] = stateSyncSnapshotInterval

	configFileOverrides["config/app.toml"] = appTomlOverrides

	// Setup
	chains := CreateThisBranchChain(t, 1, 1)
	_, ctx, _, _ := BuildInitialChain(t, chains)

	chain := chains[0].(*cosmos.CosmosChain)

	// Wait for blocks so that nodes have a few state sync snapshot available
	require.NoError(t, testutil.WaitForBlocks(ctx, stateSyncSnapshotInterval*2, chain))

	latestHeight, err := chain.Height(ctx)
	require.NoError(t, err, "failed to fetch latest chain height")

	// Trusted height should be state sync snapshot interval blocks ago.
	trustHeight := int64(latestHeight) - stateSyncSnapshotInterval

	firstFullNode := chain.FullNodes[0]

	// Fetch block hash for trusted height.
	blockRes, err := firstFullNode.Client.Block(ctx, &trustHeight)
	require.NoError(t, err, "failed to fetch trusted block")
	trustHash := hex.EncodeToString(blockRes.BlockID.Hash)

	// Construct statesync parameters for new node to get in sync.
	configFileOverrides = make(map[string]any)
	configTomlOverrides := make(testutil.Toml)

	// Set trusted parameters and rpc servers for verification.
	stateSync = make(testutil.Toml)
	stateSync["trust_hash"] = trustHash
	stateSync["trust_height"] = trustHeight
	// State sync requires minimum of two RPC servers for verification. We can provide the same RPC twice though.
	stateSync["rpc_servers"] = fmt.Sprintf("tcp://%s:26657,tcp://%s:26657", firstFullNode.HostName(), firstFullNode.HostName())
	configTomlOverrides["statesync"] = stateSync

	configFileOverrides["config/config.toml"] = configTomlOverrides

	// Now that nodes are providing state sync snapshots, state sync a new node.
	require.NoError(t, chain.AddFullNodes(ctx, configFileOverrides, 1))

	// Wait for new node to be in sync.
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	require.NoError(t, testutil.WaitForInSync(ctx, chain, chain.FullNodes[len(chain.FullNodes)-1]))
}
