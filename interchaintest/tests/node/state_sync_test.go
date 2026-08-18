package node_test

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/testutil"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"
	votingsnapshottypes "github.com/CosmosContracts/juno/v31/x/voting-snapshot/types"
)

const (
	stateSyncSnapshotInterval = 10
	stateSyncTimeout          = 3 * time.Minute
)

type NodeTestSuite struct {
	*e2esuite.E2ETestSuite
}

// stateSyncSpec returns an isolated topology. In particular, it does not add
// overrides to suite.DefaultConfig or change the integer pointers in
// suite.DefaultSpec, both of which are shared by parallel E2E packages.
func stateSyncSpec() *interchaintest.ChainSpec {
	numValidators := 1
	numFullNodes := 2 // CometBFT requires two RPC entries; use distinct live nodes.
	noHostMount := e2esuite.DefaultNoHostMount

	config := e2esuite.DefaultConfig.Clone()
	config.ConfigFileOverrides = snapshotConfigOverrides()

	return &interchaintest.ChainSpec{
		ChainName:     e2esuite.DefaultSpec.ChainName,
		Name:          e2esuite.DefaultSpec.Name,
		NumValidators: &numValidators,
		NumFullNodes:  &numFullNodes,
		Version:       e2esuite.DefaultSpec.Version,
		NoHostMount:   &noHostMount,
		ChainConfig:   config,
	}
}

func snapshotConfigOverrides() map[string]any {
	return map[string]any{
		"config/app.toml": testutil.Toml{
			"state-sync": testutil.Toml{
				"snapshot-interval":    stateSyncSnapshotInterval,
				"snapshot-keep-recent": 2,
			},
			"pruning":             "custom",
			"pruning-keep-recent": stateSyncSnapshotInterval,
			"pruning-interval":    stateSyncSnapshotInterval,
		},
	}
}

func stateSyncNodeOverrides(trustHeight int64, trustHash string, providerHosts []string) map[string]any {
	rpcServers := make([]string, len(providerHosts))
	for i, host := range providerHosts {
		rpcServers[i] = fmt.Sprintf("tcp://%s:26657", host)
	}

	return map[string]any{
		"config/config.toml": testutil.Toml{
			"statesync": testutil.Toml{
				"enable":                true,
				"rpc_servers":           strings.Join(rpcServers, ","),
				"trust_height":          trustHeight,
				"trust_hash":            trustHash,
				"trust_period":          "1h",
				"discovery_time":        "15s",
				"chunk_request_timeout": "10s",
			},
		},
	}
}

func TestNodeTestSuite(t *testing.T) {
	s := e2esuite.NewE2ETestSuite(
		[]*interchaintest.ChainSpec{stateSyncSpec()},
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

	require.Len(t, s.Chain.Validators, 1, "state-sync suite must start a snapshot validator")
	require.Len(t, s.Chain.FullNodes, 2, "state-sync suite must start two independent RPC providers")
	providers := []*cosmos.ChainNode{s.Chain.Validators[0], s.Chain.FullNodes[0]}
	providerHosts := []string{providers[0].HostName(), providers[1].HostName()}

	// Snapshot metadata uses the live application database, so querying it from
	// a second process races the running node's database lock. Advancing two
	// intervals and then successfully state-syncing the new node proves that a
	// complete snapshot was produced and served.
	require.NoError(t, testutil.WaitForBlocks(s.Ctx, stateSyncSnapshotInterval*2, s.Chain))
	latestHeight, err := s.Chain.Height(s.Ctx)
	require.NoError(t, err)
	trustHeight := latestHeight - stateSyncSnapshotInterval
	blockRes, err := providers[0].Client.Block(s.Ctx, &trustHeight)
	require.NoError(t, err,
		"trusted block query failed: provider=%s trust_height=%d latest_height=%d",
		providerHosts[0], trustHeight, latestHeight,
	)
	trustHash := strings.ToUpper(hex.EncodeToString(blockRes.BlockID.Hash))
	require.NotEmpty(t, trustHash, "empty trust hash: provider=%s trust_height=%d", providerHosts[0], trustHeight)

	t.Logf("state-sync diagnostics: providers=%v latest_height=%d trust_height=%d trust_hash=%s",
		providerHosts, latestHeight, trustHeight, trustHash)

	require.NoError(t, s.Chain.AddFullNodes(s.Ctx, stateSyncNodeOverrides(trustHeight, trustHash, providerHosts), 1),
		"add state-sync node failed: providers=%v latest_height=%d trust_height=%d trust_hash=%s",
		providerHosts, latestHeight, trustHeight, trustHash,
	)
	stateSyncNode := s.Chain.FullNodes[len(s.Chain.FullNodes)-1]

	var providerHeight, syncedHeight int64
	var providerHeightErr, syncedHeightErr error
	require.Eventually(t, func() bool {
		providerHeight, providerHeightErr = providers[0].Height(s.Ctx)
		syncedHeight, syncedHeightErr = stateSyncNode.Height(s.Ctx)
		return providerHeightErr == nil && syncedHeightErr == nil && syncedHeight >= providerHeight-1
	}, stateSyncTimeout, time.Second,
		"state-sync node did not catch tip: node=%s node_height=%d node_error=%v provider=%s provider_height=%d provider_error=%v trust_height=%d trust_hash=%s rpc_providers=%v",
		stateSyncNode.HostName(), syncedHeight, syncedHeightErr, providerHosts[0], providerHeight, providerHeightErr,
		trustHeight, trustHash, providerHosts,
	)

	// Catching the tip proves liveness, not restored-state correctness. Compare
	// one exact app hash and query consensus-sensitive module state through the
	// new node's own gRPC connection.
	verifyHeight := syncedHeight
	providerBlock, err := providers[0].Client.Block(s.Ctx, &verifyHeight)
	require.NoError(t, err)
	syncedBlock, err := stateSyncNode.Client.Block(s.Ctx, &verifyHeight)
	require.NoError(t, err)
	require.Equal(t, providerBlock.Block.Header.AppHash, syncedBlock.Block.Header.AppHash)

	expectedParams := s.QueryVotingSnapshotParams()
	expectedTotal := s.QueryTotalVotingPowerAt(verifyHeight)
	syncedVotingClient := votingsnapshottypes.NewQueryClient(stateSyncNode.GrpcConn)
	paramsResp, err := syncedVotingClient.Params(s.Ctx, &votingsnapshottypes.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, expectedParams, paramsResp.Params)
	totalResp, err := syncedVotingClient.TotalVotingPowerAt(s.Ctx, &votingsnapshottypes.QueryTotalVotingPowerAtRequest{
		AtHeight: verifyHeight,
	})
	require.NoError(t, err)
	require.Equal(t, expectedTotal, totalResp.Power)

	t.Logf("state-sync verified: initial_height=%d trust_height=%d verified_height=%d app_hash=%X voting_power=%s",
		latestHeight, trustHeight, verifyHeight, syncedBlock.Block.Header.AppHash, totalResp.Power)
}
