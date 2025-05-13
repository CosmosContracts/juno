package interchaintest

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"

	"github.com/stretchr/testify/require"

	buildertypes "github.com/skip-mev/pob/x/builder/types"
)

const (
	firstUpgradeName  = "v28"
	secondUpgradeName = "v29"
)

// v27Chain is the first version of juno that will be upgraded from
var v27Chain = ibc.DockerImage{
	Repository: JunoMainRepo,
	Version:    "v27.0.0",
	UIDGID:     "1025:1025",
}

// v28Chain is the second version of juno that will be upgraded from
var v28Chain = ibc.DockerImage{
	Repository: JunoMainRepo,
	Version:    "v28.0.2",
	UIDGID:     "1025:1025",
}

func TestFixRemovedMsgTypeQueryPanic(t *testing.T) {
	repo, localVersion := GetDockerImageInfo()
	SimulateQueryPanic(t, chainName, v28Chain.Version, localVersion, repo, firstUpgradeName, secondUpgradeName)
}

func SimulateQueryPanic(t *testing.T, chainName, firstUpgradeBranchVersion, secondUpgradeBranchVersion, upgradeRepo, firstUpgradeName, secondUpgradeName string) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	t.Parallel()

	t.Log(chainName, firstUpgradeBranchVersion, upgradeRepo, firstUpgradeName)

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
	cfg.Images = []ibc.DockerImage{v27Chain}

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

	// v27 params update
	proposalID := SubmitBuilderParamsUpdate(t, ctx, chain, chainUser)
	proposalIDInt, err := strconv.ParseUint(proposalID, 10, 64)
	VoteOnProp(t, ctx, chain, proposalIDInt, 0)
	t.Log("v27 params update proposal ID", proposalIDInt)

	// upgrade to v28
	height, err := chain.Height(ctx)
	require.NoError(t, err, "error fetching height before submit upgrade proposal")
	haltHeight := height + haltHeightDelta
	proposalID = SubmitUpgradeProposal(t, ctx, chain, chainUser, firstUpgradeName, haltHeight)
	proposalIDInt, err = strconv.ParseUint(proposalID, 10, 64)
	require.NoError(t, err, "failed to parse proposal ID")
	ValidatorVoting(t, ctx, chain, proposalIDInt, height, haltHeight)
	UpgradeNodes(t, ctx, chain, client, haltHeight, upgradeRepo, firstUpgradeBranchVersion)
	t.Log("v28 upgrade successful!")

	// upgrade to v29
	height, err = chain.Height(ctx)
	require.NoError(t, err, "error fetching height before submit upgrade proposal")
	haltHeight = height + haltHeightDelta
	proposalID = SubmitUpgradeProposal(t, ctx, chain, chainUser, secondUpgradeName, haltHeight)
	proposalIDInt, err = strconv.ParseUint(proposalID, 10, 64)
	require.NoError(t, err, "failed to parse proposal ID")
	ValidatorVoting(t, ctx, chain, proposalIDInt, height, haltHeight)
	UpgradeNodes(t, ctx, chain, client, haltHeight, upgradeRepo, secondUpgradeBranchVersion)
	t.Log("v29 upgrade successful!")

	// query gov module to check for panic
	proposals, err := chain.GovQueryProposalsV1(ctx, govv1types.ProposalStatus_PROPOSAL_STATUS_PASSED)
	require.NoError(t, err, "error querying gov module")
	t.Log("proposals", proposals)
}

func SubmitBuilderParamsUpdate(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet) string {
	govModule := sdk.MustAccAddressFromBech32("juno10d07y265gmmuvt4z0w9aw880jnsr700jvss730")

	updateParamsMsg := []cosmos.ProtoMessage{
		&buildertypes.MsgUpdateParams{
			// gov module account
			Authority: "juno10d07y265gmmuvt4z0w9aw880jnsr700jvss730",
			Params: buildertypes.Params{
				FrontRunningProtection: true,
				ProposerFee:            sdkmath.LegacyMustNewDecFromStr("1"),
				ReserveFee:             sdk.NewCoin("ujuno", sdkmath.NewInt(1)),
				MinBidIncrement:        sdk.NewCoin("ujuno", sdkmath.NewInt(1000)),
				MaxBundleSize:          100,
				EscrowAccountAddress:   govModule.Bytes(),
			},
		},
	}

	proposal, err := chain.BuildProposal(
		updateParamsMsg,
		"Update Builder Params",
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

func VoteOnProp(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, proposalID uint64, height int64) {
	err := chain.VoteOnProposalAllValidators(ctx, proposalID, cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	_, err = cosmos.PollForProposalStatus(ctx, chain, height, height+10, proposalID, govtypes.StatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")

	_, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*45)
	defer timeoutCtxCancel()
}
