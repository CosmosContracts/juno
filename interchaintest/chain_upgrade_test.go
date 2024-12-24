package interchaintest

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	junoconformance "github.com/CosmosContracts/juno/tests/interchaintest/conformance"
	helpers "github.com/CosmosContracts/juno/tests/interchaintest/helpers"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	cosmosproto "github.com/cosmos/gogoproto/proto"
	"github.com/docker/docker/client"
	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/require"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

const (
	chainName   = "juno"
	upgradeName = "v26"

	haltHeightDelta    = int64(9) // will propose upgrade this many blocks in the future
	blocksAfterUpgrade = int64(7)
)

var (
	// baseChain is the current version of the chain that will be upgraded from
	baseChain = ibc.DockerImage{
		Repository: JunoMainRepo,
		Version:    "v25.0.0",
		UidGid:     "1025:1025",
	}
)

func TestBasicJunoUpgrade(t *testing.T) {
	repo, version := GetDockerImageInfo()
	CosmosChainUpgradeTest(t, chainName, version, repo, upgradeName)
}

func CosmosChainUpgradeTest(t *testing.T, chainName, upgradeBranchVersion, upgradeRepo, upgradeName string) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	t.Parallel()

	t.Log(chainName, upgradeBranchVersion, upgradeRepo, upgradeName)

	previousVersionGenesis := []cosmos.GenesisKV{
		{
			Key:   "app_state.gov.params.voting_period",
			Value: VotingPeriod,
		},
		{
			Key:   "app_state.gov.params.max_deposit_period",
			Value: MaxDepositPeriod,
		},
		{
			Key:   "app_state.gov.params.min_deposit.0.denom",
			Value: Denom,
		},
	}

	cfg := junoConfig
	cfg.ModifyGenesis = cosmos.ModifyGenesis(previousVersionGenesis)
	cfg.Images = []ibc.DockerImage{baseChain}

	numVals, numNodes := 4, 0
	chains := CreateChainWithCustomConfig(t, numVals, numNodes, cfg)
	chain := chains[0].(*cosmos.CosmosChain)

	ic, ctx, client, _ := BuildInitialChain(t, chains)

	t.Cleanup(func() {
		_ = ic.Close()
	})

	userFunds := sdkmath.NewInt(10_000_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, chain)
	chainUser := users[0]

	// execute a contract before the upgrade
	beforeContract := junoconformance.StdExecute(t, ctx, chain, chainUser)

	// upgrade
	height, err := chain.Height(ctx)
	require.NoError(t, err, "error fetching height before submit upgrade proposal")

	haltHeight := height + haltHeightDelta
	proposalID := SubmitUpgradeProposal(t, ctx, chain, chainUser, upgradeName, haltHeight)

	proposalIDInt, err := strconv.ParseInt(proposalID, 10, 64)
	require.NoError(t, err, "failed to parse proposal ID")

	ValidatorVoting(t, ctx, chain, proposalIDInt, height, haltHeight)

	UpgradeNodes(t, ctx, chain, client, haltHeight, upgradeRepo, upgradeBranchVersion)

	// confirm we can execute against the beforeContract (ref: v20 upgrade patch)
	helpers.ExecuteMsgWithFee(t, ctx, chain, chainUser, beforeContract, "", "10000"+chain.Config().Denom, `{"increment":{}}`)

	// Post Upgrade: Conformance Validation
	junoconformance.ConformanceCosmWasm(t, ctx, chain, chainUser)
	// TODO: ibc conformance test
}

func UpgradeNodes(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, client *client.Client, haltHeight int64, upgradeRepo, upgradeBranchVersion string) {
	// bring down nodes to prepare for upgrade
	t.Log("stopping node(s)")
	err := chain.StopAllNodes(ctx)
	require.NoError(t, err, "error stopping node(s)")

	// upgrade version on all nodes
	t.Log("upgrading node(s)")
	chain.UpgradeVersion(ctx, client, upgradeRepo, upgradeBranchVersion)

	// start all nodes back up.
	// validators reach consensus on first block after upgrade height
	// and chain block production resumes.
	t.Log("starting node(s)")
	err = chain.StartAllNodes(ctx)
	require.NoError(t, err, "error starting upgraded node(s)")

	timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*60)
	defer timeoutCtxCancel()

	err = testutil.WaitForBlocks(timeoutCtx, int(blocksAfterUpgrade), chain)
	require.NoError(t, err, "chain did not produce blocks after upgrade")

	height, err := chain.Height(ctx)
	require.NoError(t, err, "error fetching height after upgrade")

	require.GreaterOrEqual(t, height, haltHeight+blocksAfterUpgrade, "height did not increment enough after upgrade")
}

func ValidatorVoting(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, proposalID int64, height int64, haltHeight int64) {
	err := chain.VoteOnProposalAllValidators(ctx, proposalID, cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	_, err = cosmos.PollForProposalStatus(ctx, chain, height, height+haltHeightDelta, proposalID, govtypes.StatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")

	timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*45)
	defer timeoutCtxCancel()

	height, err = chain.Height(ctx)
	require.NoError(t, err, "error fetching height before upgrade")

	// this should timeout due to chain halt at upgrade height.
	_ = testutil.WaitForBlocks(timeoutCtx, int(haltHeight-height), chain)

	height, err = chain.Height(ctx)
	require.NoError(t, err, "error fetching height after chain should have halted")

	// make sure that chain is halted
	require.Equal(t, haltHeight, height, "height is not equal to halt height")
}

func SubmitUpgradeProposal(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, upgradeName string, haltHeight int64) string {
	upgradeMsg := []cosmosproto.Message{
		&upgradetypes.MsgSoftwareUpgrade{
			// gGov Module account
			Authority: "juno10d07y265gmmuvt4z0w9aw880jnsr700jvss730",
			Plan: upgradetypes.Plan{
				Name:   upgradeName,
				Height: int64(haltHeight),
			},
		},
	}

	proposal, err := chain.BuildProposal(upgradeMsg, "Chain Upgrade 1", "Summary desc", "ipfs://CID", fmt.Sprintf(`500000000%s`, chain.Config().Denom))
	require.NoError(t, err, "error building proposal")

	txProp, err := chain.SubmitProposal(ctx, user.KeyName(), proposal)
	t.Log("txProp", txProp)
	require.NoError(t, err, "error submitting proposal")

	return txProp.ProposalID
}
