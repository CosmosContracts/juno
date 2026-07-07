package suite

import (
	"context"
	"strings"
	"time"

	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/testutil"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/require"
)

var blocksAfterUpgrade = int64(7)

func (s *E2ETestSuite) UpgradeNodes(chain *cosmos.CosmosChain, client *client.Client, haltHeight int64, upgradeRepo, upgradeBranchVersion string) {
	t := s.T()
	// bring down nodes to prepare for upgrade
	t.Log("stopping node(s)")
	err := chain.StopAllNodes(s.Ctx)
	require.NoError(t, err, "error stopping node(s)")

	// upgrade version on all nodes
	t.Log("upgrading node(s)")
	chain.UpgradeVersion(s.Ctx, client, upgradeRepo, upgradeBranchVersion)

	// Start all nodes back up on the upgraded image; validators reach
	// consensus on the first block after the upgrade height and block
	// production resumes.
	//
	// StartAllNodes recreates each node container and then, in the window
	// right after `docker start`, reads the container's mapped host RPC port.
	// On a warm recreation (post-upgrade) the daemon often hasn't finished
	// *publishing* the port when interchaintest reads it, so GetHostPorts
	// returns "" and StartContainer dials "tcp://" — failing with
	// `Post "http:": http: no Host in request URL` after its internal retry.
	// The container and its published port are healthy a moment later
	// (verified via docker inspect: every node reaches `running exit=0` with
	// 26657 published); the framework just read the mapping too early and
	// cached the empty value. We let StartAllNodes run to completion — so the
	// containers are fully settled rather than interrupted mid-startup — and
	// then unconditionally rebuild every node's client from the live,
	// now-published port mapping. Rebuilding is idempotent: on the (common)
	// happy path it re-reads the same ports StartAllNodes already resolved; on
	// the racy path it repairs the clients StartAllNodes left dialing "tcp://".
	t.Log("starting node(s)")
	if startErr := chain.StartAllNodes(s.Ctx); startErr != nil {
		t.Logf("StartAllNodes reported %v; recovering published host ports from the live containers", startErr)
	}
	s.rebuildNodeClientsAfterRestart(chain)

	timeoutCtx, timeoutCtxCancel := context.WithTimeout(s.Ctx, time.Second*60)
	defer timeoutCtxCancel()

	err = testutil.WaitForBlocks(timeoutCtx, int(blocksAfterUpgrade), chain)
	require.NoError(t, err, "chain did not produce blocks after upgrade")

	height, err := chain.Height(s.Ctx)
	require.NoError(t, err, "error fetching height after upgrade")

	require.GreaterOrEqual(t, height, haltHeight+blocksAfterUpgrade, "height did not increment enough after upgrade")

	// UpgradeVersion recreated every node container, so the host gRPC port
	// changed. Re-dial before any post-upgrade gRPC query, or they all fail
	// with "connection refused" against the stale pre-upgrade port.
	s.RefreshGRPCClients()
}

// rebuildNodeClientsAfterRestart recovers from interchaintest's host-port
// publish race (see UpgradeNodes): after a container recreation the framework
// may cache an empty host RPC port, leaving each node's Tendermint client
// dialing "tcp://". The container is running and its port is published a
// moment later, so for every node we wait for the daemon to expose the mapped
// 26657 port (GetHostAddress re-inspects the live container) and rebuild the
// node's RPC client from the real address. NewClient also re-dials the node's
// gRPC conn from the still-empty cached port, but nothing in the suite uses
// that per-node conn — queries go through the suite's own client, refreshed
// separately by RefreshGRPCClients.
func (s *E2ETestSuite) rebuildNodeClientsAfterRestart(chain *cosmos.CosmosChain) {
	t := s.T()
	for _, node := range chain.Nodes() {
		var rpcAddr string
		s.Require().Eventuallyf(func() bool {
			addr, err := node.GetHostAddress(s.Ctx, "26657/tcp")
			if err != nil {
				return false
			}
			// GetHostAddress returns "http://host:port"; NewClient wants a
			// tcp:// URL (it maps the scheme back to http internally).
			rpcAddr = "tcp://" + strings.TrimPrefix(addr, "http://")
			return true
		}, 60*time.Second, time.Second, "node %s RPC port was never published after restart", node.Name())

		require.NoError(t, node.NewClient(rpcAddr), "failed to rebuild RPC client for node %s", node.Name())
	}
}
