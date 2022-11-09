//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"time"

	appparams "github.com/CosmosContracts/juno/v11/app/params"
	"github.com/CosmosContracts/juno/v11/tests/e2e/initialization"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestIBCTokenTransfer tests that IBC token transfers work as expected.
// Additionally, it attempst to create a pool with IBC denoms.
func (s *IntegrationTestSuite) TestIBCTokenTransferAndCreatePool() {
	if s.skipIBC {
		s.T().Skip("Skipping IBC tests")
	}
	chainA := s.configurer.GetChainConfig(0)
	chainB := s.configurer.GetChainConfig(1)
	chainA.SendIBC(chainB, chainB.NodeConfigs[0].PublicAddress, initialization.JunoToken)
	chainB.SendIBC(chainA, chainA.NodeConfigs[0].PublicAddress, initialization.JunoToken)
	chainA.SendIBC(chainB, chainB.NodeConfigs[0].PublicAddress, initialization.StakeToken)
	chainB.SendIBC(chainA, chainA.NodeConfigs[0].PublicAddress, initialization.StakeToken)

}

//TODO

// func (s *IntegrationTestSuite) TestStateSync() {
// 	if s.skipStateSync {
// 		s.T().Skip()
// 	}

// 	chainA := s.configurer.GetChainConfig(0)
// 	runningNode, err := chainA.GetDefaultNode()
// 	s.Require().NoError(err)

// 	persistentPeers := chainA.GetPersistentPeers()

// 	stateSyncHostPort := fmt.Sprintf("%s:26657", runningNode.Name)
// 	stateSyncRPCServers := []string{stateSyncHostPort, stateSyncHostPort}

// 	// get trust height and trust hash.
// 	trustHeight, err := runningNode.QueryCurrentHeight()
// 	s.Require().NoError(err)

// 	trustHash, err := runningNode.QueryHashFromBlock(trustHeight)
// 	s.Require().NoError(err)

// 	stateSynchingNodeConfig := &initialization.NodeConfig{
// 		Name:               "state-sync",
// 		Pruning:            "default",
// 		PruningKeepRecent:  "0",
// 		PruningInterval:    "0",
// 		SnapshotInterval:   1500,
// 		SnapshotKeepRecent: 2,
// 	}

// 	tempDir, err := os.MkdirTemp("", "e2e-statesync-")
// 	s.Require().NoError(err)

// 	// configure genesis and config files for the state-synchin node.
// 	nodeInit, err := initialization.InitSingleNode(
// 		chainA.Id,
// 		tempDir,
// 		filepath.Join(runningNode.ConfigDir, "config", "genesis.json"),
// 		stateSynchingNodeConfig,
// 		time.Duration(chainA.VotingPeriod),
// 		// time.Duration(chainA.ExpeditedVotingPeriod),
// 		trustHeight,
// 		trustHash,
// 		stateSyncRPCServers,
// 		persistentPeers,
// 	)
// 	s.Require().NoError(err)

// 	stateSynchingNode := chainA.CreateNode(nodeInit)

// 	// ensure that the running node has snapshots at a height > trustHeight.
// 	hasSnapshotsAvailable := func(syncInfo coretypes.SyncInfo) bool {
// 		snapshotHeight := runningNode.SnapshotInterval
// 		if uint64(syncInfo.LatestBlockHeight) < snapshotHeight {
// 			s.T().Logf("snapshot height is not reached yet, current (%d), need (%d)", syncInfo.LatestBlockHeight, snapshotHeight)
// 			return false
// 		}

// 		snapshots, err := runningNode.QueryListSnapshots()
// 		s.Require().NoError(err)

// 		for _, snapshot := range snapshots {
// 			if snapshot.Height > uint64(trustHeight) {
// 				s.T().Log("found state sync snapshot after trust height")
// 				return true
// 			}
// 		}
// 		s.T().Log("state sync snashot after trust height is not found")
// 		return false
// 	}
// 	runningNode.WaitUntil(hasSnapshotsAvailable)

// 	// start the state synchin node.
// 	err = stateSynchingNode.Run()
// 	s.NoError(err)

// 	// ensure that the state synching node cathes up to the running node.
// 	s.Require().Eventually(func() bool {
// 		stateSyncNodeHeight, err := stateSynchingNode.QueryCurrentHeight()
// 		s.Require().NoError(err)
// 		runningNodeHeight, err := runningNode.QueryCurrentHeight()
// 		s.Require().NoError(err)
// 		return stateSyncNodeHeight == runningNodeHeight
// 	},
// 		3*time.Minute,
// 		500*time.Millisecond,
// 	)

// 	// stop the state synching node.
// 	err = chainA.RemoveNode(stateSynchingNode.Name)
// 	s.NoError(err)
// }

func (s *IntegrationTestSuite) TestExpeditedProposals() {
	chainA := s.configurer.GetChainConfig(0)
	chainANode, err := chainA.GetDefaultNode()
	s.NoError(err)

	chainANode.SubmitTextProposal("expedited text proposal", sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(1000000)))
	chainA.LatestProposalNumber += 1
	chainANode.DepositProposal(chainA.LatestProposalNumber, true)
	totalTimeChan := make(chan time.Duration, 1)
	go chainANode.QueryPropStatusTimed(chainA.LatestProposalNumber, "PROPOSAL_STATUS_VOTING_PERIOD", totalTimeChan)
	for _, node := range chainA.NodeConfigs {
		node.VoteYesProposal(initialization.ValidatorWalletName, chainA.LatestProposalNumber)
	}
	// if querying proposal takes longer than timeoutPeriod, stop the goroutine and error
	var elapsed time.Duration
	timeoutPeriod := time.Duration(2 * time.Minute)
	select {
	case elapsed = <-totalTimeChan:
	case <-time.After(timeoutPeriod):
		err := fmt.Errorf("go routine took longer than %s", timeoutPeriod)
		s.Require().NoError(err)
	}

	// compare the time it took to reach pass status to expected expedited voting period
	expeditedVotingPeriodDuration := time.Duration(chainA.ExpeditedVotingPeriod * float32(time.Second))
	timeDelta := elapsed - expeditedVotingPeriodDuration
	// ensure delta is within two seconds of expected time
	s.Require().Less(timeDelta, 3*time.Second)
	s.T().Logf("expeditedVotingPeriodDuration within two seconds of expected time: %v", timeDelta)
	close(totalTimeChan)
}
