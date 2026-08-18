package node_test

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
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

var snapshotLineRE = regexp.MustCompile(`height:\s*(\d+)\s+format:\s*(\d+)\s+chunks:\s*(\d+)`)

type snapshotMetadata struct {
	Height uint64
	Format uint32
	Chunks uint32
}

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
	// Pass the SDK flags too: they are the authoritative app-options keys used
	// when the snapshot manager is constructed.
	config.AdditionalStartArgs = append(config.AdditionalStartArgs,
		"--state-sync.snapshot-interval", "10",
		"--state-sync.snapshot-keep-recent", "2",
	)

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
			// The snapshot interval must be a multiple of pruning-keep-every.
			"pruning":             "custom",
			"pruning-keep-recent": stateSyncSnapshotInterval,
			"pruning-keep-every":  stateSyncSnapshotInterval,
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

func parseSnapshotMetadata(output []byte) ([]snapshotMetadata, error) {
	matches := snapshotLineRE.FindAllSubmatch(output, -1)
	snapshots := make([]snapshotMetadata, 0, len(matches))
	for _, match := range matches {
		height, err := strconv.ParseUint(string(match[1]), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse snapshot height: %w", err)
		}
		format, err := strconv.ParseUint(string(match[2]), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("parse snapshot format: %w", err)
		}
		chunks, err := strconv.ParseUint(string(match[3]), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("parse snapshot chunks: %w", err)
		}
		snapshots = append(snapshots, snapshotMetadata{Height: height, Format: uint32(format), Chunks: uint32(chunks)})
	}
	return snapshots, nil
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

	require.Len(t, s.Chain.FullNodes, 2, "state-sync suite must start two snapshot/RPC providers")
	providers := []*cosmos.ChainNode{s.Chain.FullNodes[0], s.Chain.FullNodes[1]}
	providerHosts := []string{providers[0].HostName(), providers[1].HostName()}

	var snapshots []snapshotMetadata
	var snapshotQueryErr error
	require.Eventually(t, func() bool {
		stdout, stderr, err := providers[0].ExecBin(s.Ctx, "snapshots", "list")
		if err != nil {
			snapshotQueryErr = fmt.Errorf("snapshots list: %w (stderr: %s)", err, strings.TrimSpace(string(stderr)))
			return false
		}
		snapshots, snapshotQueryErr = parseSnapshotMetadata(stdout)
		return snapshotQueryErr == nil && len(snapshots) > 0 && snapshots[0].Height > stateSyncSnapshotInterval && snapshots[0].Chunks > 0
	}, 90*time.Second, 2*time.Second,
		"no usable snapshot from provider host=%s: snapshots=%+v count=%d query_error=%v",
		providerHosts[0], snapshots, len(snapshots), snapshotQueryErr,
	)

	// Anchor below the newest snapshot so the light client can verify the
	// snapshot and all subsequent blocks. The block hash comes from a provider,
	// not from locally derived state.
	newestSnapshot := snapshots[0]
	trustHeight := int64(newestSnapshot.Height) - stateSyncSnapshotInterval
	blockRes, err := providers[0].Client.Block(s.Ctx, &trustHeight)
	require.NoError(t, err,
		"trusted block query failed: provider=%s trust_height=%d snapshot=%+v snapshot_count=%d",
		providerHosts[0], trustHeight, newestSnapshot, len(snapshots),
	)
	trustHash := strings.ToUpper(hex.EncodeToString(blockRes.BlockID.Hash))
	require.NotEmpty(t, trustHash, "empty trust hash: provider=%s trust_height=%d", providerHosts[0], trustHeight)

	t.Logf("state-sync diagnostics: providers=%v trust_height=%d trust_hash=%s snapshots=%+v snapshot_count=%d",
		providerHosts, trustHeight, trustHash, snapshots, len(snapshots))

	require.NoError(t, s.Chain.AddFullNodes(s.Ctx, stateSyncNodeOverrides(trustHeight, trustHash, providerHosts), 1),
		"add state-sync node failed: providers=%v trust_height=%d trust_hash=%s snapshot=%+v snapshot_count=%d",
		providerHosts, trustHeight, trustHash, newestSnapshot, len(snapshots),
	)
	stateSyncNode := s.Chain.FullNodes[len(s.Chain.FullNodes)-1]

	var providerHeight, syncedHeight int64
	var providerHeightErr, syncedHeightErr error
	require.Eventually(t, func() bool {
		providerHeight, providerHeightErr = providers[0].Height(s.Ctx)
		syncedHeight, syncedHeightErr = stateSyncNode.Height(s.Ctx)
		return providerHeightErr == nil && syncedHeightErr == nil && syncedHeight >= providerHeight-1
	}, stateSyncTimeout, time.Second,
		"state-sync node did not catch tip: node=%s node_height=%d node_error=%v provider=%s provider_height=%d provider_error=%v trust_height=%d trust_hash=%s snapshot=%+v snapshot_count=%d rpc_providers=%v",
		stateSyncNode.HostName(), syncedHeight, syncedHeightErr, providerHosts[0], providerHeight, providerHeightErr,
		trustHeight, trustHash, newestSnapshot, len(snapshots), providerHosts,
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

	t.Logf("state-sync verified: snapshot_height=%d trust_height=%d verified_height=%d app_hash=%X voting_power=%s",
		newestSnapshot.Height, trustHeight, verifyHeight, syncedBlock.Block.Header.AppHash, totalResp.Power)
}
