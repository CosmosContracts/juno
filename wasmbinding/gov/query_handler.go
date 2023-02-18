package gov

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/CosmosContracts/juno/v13/wasmbinding/gov/types"
)

// KeeperReaderExpected defines the x/gov keeper expected read operations.
type KeeperReaderExpected interface {
	GetVote(c sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) (vote govTypes.Vote, found bool)
}

// QueryHandler provides a custom WASM query handler for the x/gov module.
type QueryHandler struct {
	govKeeper KeeperReaderExpected
}

// NewQueryHandler creates a new QueryHandler instance.
func NewQueryHandler(gk KeeperReaderExpected) QueryHandler {
	return QueryHandler{
		govKeeper: gk,
	}
}

// GetVote returns the vote weighted options for a given proposal and voter.
func (h QueryHandler) GetVote(ctx sdk.Context, req types.VoteRequest) (types.VoteResponse, error) {
	if err := req.Validate(); err != nil {
		return types.VoteResponse{}, fmt.Errorf("vote: %w", err)
	}

	vote, found := h.govKeeper.GetVote(ctx, req.ProposalID, req.MustGetVoter())
	if !found {
		err := sdkErrors.Wrap(govTypes.ErrInvalidVote, fmt.Errorf("vote not found for proposal %d and voter %s", req.ProposalID, req.Voter).Error())
		return types.VoteResponse{}, err
	}

	return types.NewVoteResponse(vote), nil
}
