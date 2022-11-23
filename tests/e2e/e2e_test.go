package main

import (
	"fmt"

	"github.com/CosmosContracts/juno/v12/tests/e2e/initialization"
)

// TestIBCTokenTransfer tests that IBC token transfers work as expected.
func (s *IntegrationTestSuite) TestIBCTokenTransfer() {
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

// TestTokenFactoryBindings tests that the TokenFactory module and its bindings work as expected.
func (s *IntegrationTestSuite) TestTokenFactoryBindings() {
	chainA := s.configurer.GetChainConfig(0)

	// get teh keyName of chainA node, using internalNode
	addr := chainA.NodeConfigs[0].PublicAddress

	chainA.NodeConfigs[0].StoreWasmCode("scripts/tokenfactory.wasm", addr) // code_id: 1
	chainA.LatestCodeId = 1
	codeId := fmt.Sprint(chainA.LatestCodeId)

	contractAddr, err := chainA.NodeConfigs[0].InstantiateWasmContract(codeId, "{}", "tokenfactorylabel", addr)
	s.Require().NoError(err)

	println("contractAddr: ", contractAddr)

	/*
		- Execute on contract to:
		- query cost to make a token (1 juno)
		- create subdenom token and make sure it cost 1 juno. Try less (err), and more (still success?)
		- mint 100 tokens
		- send 50 tokens to another account
		- query balances and ensure each have 50 tokens

		- Burn 10 from contract (success)
		- Burn 10 from another account (fail)

		- Try to create a token with a name that already exists (fail)
		- Try to create a token with a name that is too long (fail)
		- Try to create a token with a name that is too short (fail)

		- change admin to other account
		- try to mint more tokens (err) from old admin
		- try to mint more tokens (success) from new admin


	*/
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
