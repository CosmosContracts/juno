//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/CosmosContracts/juno/v12/tests/e2e/initialization"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	tfSuccessCode  = "\"code\":0,"
	tfAlreadyExist = "attempting to create a denom that already exists"
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
// docker network prune && make test-e2e-skip
func (s *IntegrationTestSuite) TestTokenFactoryBindings() {
	chainA := s.configurer.GetChainConfig(0)
	node := chainA.NodeConfigs[0]
	wallet := initialization.ValidatorWalletName

	params, err := node.QueryTokenFactoryParams()
	s.Require().NoError(err)
	mintCost := params.Params.DenomCreationFee[0]

	mintCostStr := fmt.Sprintf("%s%s", mintCost.Amount.String(), mintCost.Denom)

	// Store Contract
	node.StoreWasmCode("/juno/tokenfactory.wasm", wallet)
	chainA.LatestCodeID = 1

	// Instantiate to codeId 1
	node.InstantiateWasmContract(
		strconv.Itoa(chainA.LatestCodeID),
		"{}",
		"tokenfactorylabel",
		wallet,
		"", // no admin
	)

	// Get codeId 1 contracts
	contracts, err := node.QueryContractsFromID(chainA.LatestCodeID)
	s.NoError(err)
	s.Require().Len(contracts, 1, "Wrong number of contracts for the tokenfactory.wasm initialization")

	contractAddr := contracts[0]

	// Successfully create a denom for the wasm contract
	node.WasmExecute(contractAddr, `{"create_denom":{"subdenom":"test"}}`, wallet, mintCostStr, tfSuccessCode)
	// failing to create a denom
	node.WasmExecute(contractAddr, fmt.Sprintf(`{"create_denom":{"subdenom":"%s"}}`, strings.Repeat("a", 61)), wallet, mintCostStr, "subdenom too long")

	ourDenom := fmt.Sprintf("factory/%s/test", contractAddr) // factory/juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8/test

	denoms, err := node.QueryDenomsFromCreator(contractAddr)
	s.NoError(err)
	s.Require().Len(denoms, 1, "Wrong number of denoms for the token factory")

	// Mint some tokens to an account (our contract in this case) via bank module
	amt := 100
	msg := fmt.Sprintf(`{"mint_tokens":{"amount":"%d","denom":"%s","mint_to_address":"%s"}}`, amt, ourDenom, contractAddr)
	node.WasmExecute(contractAddr, msg, wallet, "", tfSuccessCode)

	// Mint Balance Check
	balance, err := node.QueryBalances(contractAddr)
	s.Require().NoError(err)
	s.checkBalance(balance, sdk.NewCoins(sdk.NewCoin(ourDenom, sdk.NewInt(int64(amt)))))

	// Burn some of the tokens (can only be done for the contract which owns them = blank)
	msg = fmt.Sprintf(`{"burn_tokens":{"amount":"5","denom":"%s","burn_from_address":""}}`, ourDenom)
	node.WasmExecute(contractAddr, msg, wallet, "", tfSuccessCode)

	// Balance Check after burn -5
	balance, err = node.QueryBalances(contractAddr)
	s.Require().NoError(err)
	s.checkBalance(balance, sdk.NewCoins(sdk.NewCoin(ourDenom, sdk.NewInt(int64(amt-5)))))

	// Transfer admin to another account
	msg = fmt.Sprintf(`{"change_admin":{"denom":"%s","new_admin_address":"%s"}}`, ourDenom, "juno1aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaavju90c")
	node.WasmExecute(contractAddr, msg, wallet, "", tfSuccessCode)
}

func (s *IntegrationTestSuite) checkBalance(coins sdk.Coins, expected sdk.Coins) {
	for _, coin := range coins {
		for _, expectedCoin := range expected {
			if coin.Denom == expectedCoin.Denom {
				s.Require().Equal(expectedCoin.Amount, coin.Amount)
			}
		}
	}
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
