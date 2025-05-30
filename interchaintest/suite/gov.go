package suite

import (
	"context"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"

	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func (s *E2ETestSuite) ValidatorVoting(chain *cosmos.CosmosChain, proposalID uint64, height int64, haltHeight int64) {
	t := s.T()
	err := chain.VoteOnProposalAllValidators(s.Ctx, proposalID, cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	_, err = cosmos.PollForProposalStatus(s.Ctx, chain, height, height+DefaultHaltHeightDelta, proposalID, govv1beta1types.StatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")

	timeoutCtx, timeoutCtxCancel := context.WithTimeout(s.Ctx, time.Second*45)
	defer timeoutCtxCancel()

	height, err = chain.Height(s.Ctx)
	require.NoError(t, err, "error fetching height before upgrade")

	// this should timeout due to chain halt at upgrade height.
	_ = testutil.WaitForBlocks(timeoutCtx, int(haltHeight-height), chain)

	height, err = chain.Height(s.Ctx)
	require.NoError(t, err, "error fetching height after chain should have halted")

	// make sure that chain is halted
	require.Equal(t, haltHeight, height, "height is not equal to halt height")
}
