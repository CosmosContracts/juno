package node_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"
)

const stateSyncSnapshotInterval = 10

type NodeTestSuite struct {
	*e2esuite.E2ETestSuite
}

func TestNodeTestSuite(t *testing.T) {
	s := e2esuite.NewE2ETestSuite(
		[]*interchaintest.ChainSpec{e2esuite.DefaultSpec},
		e2esuite.DefaultTxCfg,
	)

	t.Parallel()
	t.Cleanup(func() {
		_ = s.Ic.Close()
	})

	testSuite := &NodeTestSuite{E2ETestSuite: s}
	suite.Run(t, testSuite)
}

func (s *NodeTestSuite) TestStateSync() {
	t := s.T()
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	configFileOverrides := make(map[string]any)
	appTomlOverrides := make(testutil.Toml)

	// state sync snapshots every stateSyncSnapshotInterval blocks.
	stateSync := make(testutil.Toml)
	stateSync["snapshot-interval"] = stateSyncSnapshotInterval
	appTomlOverrides["state-sync"] = stateSync

	// state sync snapshot interval must be a multi^ple of pruning keep every interval.
	appTomlOverrides["pruning"] = "custom"
	appTomlOverrides["pruning-keep-recent"] = stateSyncSnapshotInterval
	appTomlOverrides["pruning-keep-every"] = stateSyncSnapshotInterval
	appTomlOverrides["pruning-interval"] = stateSyncSnapshotInterval

	configFileOverrides["config/app.toml"] = appTomlOverrides

	// Wait for blocks so that nodes have a few state sync snapshot available
	require.NoError(t, testutil.WaitForBlocks(s.Ctx, stateSyncSnapshotInterval*2, s.Chain))

	latestHeight, err := s.Chain.Height(s.Ctx)
	require.NoError(t, err, "failed to fetch latest chain height")

	// Trusted height should be state sync snapshot interval blocks ago.
	trustHeight := int64(latestHeight) - stateSyncSnapshotInterval

	firstFullNode := s.Chain.FullNodes[0]

	// Fetch block hash for trusted height.
	blockRes, err := firstFullNode.Client.Block(s.Ctx, &trustHeight)
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
	require.NoError(t, s.Chain.AddFullNodes(s.Ctx, configFileOverrides, 1))

	// Wait for new node to be in sync.
	ctx, cancel := context.WithTimeout(s.Ctx, 30*time.Second)
	defer cancel()
	require.NoError(t, testutil.WaitForInSync(ctx, s.Chain, s.Chain.FullNodes[len(s.Chain.FullNodes)-1]))
}
