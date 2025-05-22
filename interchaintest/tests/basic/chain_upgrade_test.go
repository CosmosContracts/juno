package basic

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/docker/docker/client"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	helpers "github.com/CosmosContracts/juno/tests/interchaintest/helpers"
	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"
)

const (
	chainName   = "juno"
	upgradeName = "v30"

	haltHeightDelta    = int64(9) // will propose upgrade this many blocks in the future
	blocksAfterUpgrade = int64(7)
)

// baseChain is the current version of the chain that will be upgraded from
var baseChain = ibc.DockerImage{
	Repository: e2esuite.JunoRepo,
	Version:    "v29.0.0",
	UIDGID:     "1025:1025",
}

type BasicUpgradeTestSuite struct {
	e2esuite.E2ETestSuite
}

func TestBasicUpgradeTestSuite(t *testing.T) {
	previousVersionGenesis := []cosmos.GenesisKV{
		{
			Key:   "app_state.gov.params.voting_period",
			Value: e2esuite.VotingPeriod,
		},
		{
			Key:   "app_state.gov.params.max_deposit_period",
			Value: e2esuite.MaxDepositPeriod,
		},
		{
			Key:   "app_state.gov.params.min_deposit.0.denom",
			Value: e2esuite.Denom,
		},
	}

	cfg := e2esuite.Config
	cfg.ModifyGenesis = cosmos.ModifyGenesis(previousVersionGenesis)
	cfg.Images = []ibc.DockerImage{baseChain}
	numValidators := 4
	numFullNodes := 1
	noHostMount := false
	spec := &interchaintest.ChainSpec{
		ChainName:     "juno-upgrade",
		Name:          "juno-upgrade",
		NumValidators: &numValidators,
		NumFullNodes:  &numFullNodes,
		Version:       baseChain.Version,
		NoHostMount:   &noHostMount,
		ChainConfig:   cfg,
	}

	s := e2esuite.NewE2ETestSuite(
		spec,
		e2esuite.TxCfg,
	)

	suite.Run(t, s)
}

func (s *BasicUpgradeTestSuite) TestBasicJunoUpgrade() {
	t := s.T()
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	t.Parallel()

	repo, version := e2esuite.GetDockerImageInfo()
	user := s.GetAndFundTestUser(s.Ctx, t.Name(), 10_000_000_000, s.Chain)

	var chains []*cosmos.CosmosChain = make([]*cosmos.CosmosChain, 0)
	chains = append(chains, s.Chain)
	_, client := s.Icc(s.Ctx, t, chains)

	// execute a contract before the upgrade
	beforeContract := e2esuite.StdExecute(t, s.Ctx, s.Chain, user)

	// upgrade
	height, err := s.Chain.Height(s.Ctx)
	require.NoError(t, err, "error fetching height before submit upgrade proposal")

	haltHeight := height + haltHeightDelta
	proposalID := SubmitUpgradeProposal(t, s.Ctx, s.Chain, user, upgradeName, haltHeight)

	proposalIDInt, err := strconv.ParseUint(proposalID, 10, 64)
	require.NoError(t, err, "failed to parse proposal ID")

	ValidatorVoting(t, s.Ctx, s.Chain, proposalIDInt, height, haltHeight)
	UpgradeNodes(t, s.Ctx, s.Chain, client, haltHeight, repo, version)

	// confirm we can execute against the beforeContract (ref: v20 upgrade patch)
	_, err = helpers.ExecuteMsgWithFeeReturn(t, s.Ctx, s.Chain, user, beforeContract, "", "10000"+s.Chain.Config().Denom, `{"increment":{}}`)
	require.NoError(t, err)

	// Post Upgrade: Conformance Validation
	e2esuite.ConformanceCosmWasm(t, s.Ctx, s.Chain, user)

	t.Cleanup(func() {
		_ = s.Ic.Close()
	})
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

func ValidatorVoting(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, proposalID uint64, height int64, haltHeight int64) {
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
	upgradeMsg := []cosmos.ProtoMessage{
		&upgradetypes.MsgSoftwareUpgrade{
			// gov module account
			Authority: "juno10d07y265gmmuvt4z0w9aw880jnsr700jvss730",
			Plan: upgradetypes.Plan{
				Name:   upgradeName,
				Height: int64(haltHeight),
			},
		},
	}

	proposal, err := chain.BuildProposal(
		upgradeMsg,
		"Chain Upgrade 1",
		"Summary desc",
		"ipfs://CID",
		fmt.Sprintf(`500000000%s`, chain.Config().Denom),
		sdk.MustBech32ifyAddressBytes("juno", user.Address()),
		false)
	require.NoError(t, err, "error building proposal")

	txProp, err := chain.SubmitProposal(ctx, user.KeyName(), proposal)
	t.Log("txProp", txProp)
	require.NoError(t, err, "error submitting proposal")

	return txProp.ProposalID
}
