package keeper

import (
	"encoding/json"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

type GovHooks struct {
	k Keeper
}

var _ govtypes.GovHooks = GovHooks{}

func (k Keeper) GovHooks() GovHooks {
	return GovHooks{k: k}
}

type Proposal struct {
	ProposalId uint64 `json:"proposal_id"`
	Proposer   string `json:"proposer"`
	Status     uint   `json:"status"`
	SubmitTime string `json:"submit_time"`
	Metadata   string `json:"metadata"`
	Title      string `json:"title"`
	Summary    string `json:"summary"`
}

func NewProposal(prop v1.Proposal) Proposal {
	return Proposal{
		ProposalId: prop.Id,
		Proposer:   prop.Proposer,
		Status:     uint(prop.Status),
		SubmitTime: strconv.Itoa(prop.SubmitTime.Second()),
		Metadata:   prop.GetMetadata(),
		Title:      prop.GetTitle(),
		Summary:    prop.GetSummary(),
	}
}

type Vote struct {
	ProposalId   uint64                   `json:"proposal_id"`
	VoterAddress string                   `json:"voter_address"`
	VoteOption   []*v1.WeightedVoteOption `json:"vote_option"` // TODO: Can we read this in cw? [{"option":1,"weight":"1.00"}]
}

func NewVote(vote v1.Vote) Vote {
	return Vote{
		ProposalId:   vote.ProposalId,
		VoterAddress: vote.Voter,
		VoteOption:   vote.Options,
	}
}

type SudoMsgAfterProposalSubmission struct {
	AfterProposalSubmission Proposal `json:"after_proposal_submission"`
}

type SudoMsgAfterProposalDeposit struct {
	AfterProposalDeposit Proposal `json:"after_proposal_deposit"`
}

type SudoMsgAfterProposalVote struct {
	AfterProposalVote Vote `json:"after_proposal_vote"`
}

type SudoAfterProposalVotingPeriodEnded struct {
	AfterProposalVotingPeriodEnded string `json:"after_proposal_voting_period_ended"`
}

// TODO: move this to the keeper, and either caLL it with "staking" or "gov" using module type names?
func (h GovHooks) sendMsgToAll(ctx sdk.Context, msgBz []byte) error {
	// on errors return nil, if in a loop continue.

	// TODO: add this in the keeper, anyone can register it.
	// iter all contracts here
	contract, err := sdk.AccAddressFromBech32("juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8")
	if err != nil {
		return nil
	}

	// 100k/250k gas limit?
	gasLimitCtx := ctx.WithGasMeter(sdk.NewGasMeter(100_000))
	if _, err = h.k.contractKeeper.Sudo(gasLimitCtx, contract, msgBz); err != nil {
		return nil
	}

	// ctx.GasMeter().ConsumeGas(100_000, "cw-hooks: AfterValidatorCreated")
	return nil
}

func (h GovHooks) AfterProposalSubmission(ctx sdk.Context, proposalID uint64) {
	prop, found := h.k.govKeeper.GetProposal(ctx, proposalID)
	if !found {
		return
	}

	msgBz, err := json.Marshal(SudoMsgAfterProposalSubmission{
		AfterProposalSubmission: NewProposal(prop),
	})
	if err != nil {
		return
	}

	h.sendMsgToAll(ctx, msgBz)
}

func (h GovHooks) AfterProposalDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress) {
	prop, found := h.k.govKeeper.GetProposal(ctx, proposalID)
	if !found {
		return
	}

	msgBz, err := json.Marshal(SudoMsgAfterProposalDeposit{
		AfterProposalDeposit: NewProposal(prop),
	})
	if err != nil {
		return
	}

	h.sendMsgToAll(ctx, msgBz)
}

func (h GovHooks) AfterProposalVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {
	vote, found := h.k.govKeeper.GetVote(ctx, proposalID, voterAddr)
	if !found {
		return
	}

	msgBz, err := json.Marshal(SudoMsgAfterProposalVote{
		AfterProposalVote: NewVote(vote),
	})
	if err != nil {
		return
	}

	h.sendMsgToAll(ctx, msgBz)
}

func (h GovHooks) AfterProposalFailedMinDeposit(ctx sdk.Context, proposalID uint64) {
}

func (h GovHooks) AfterProposalVotingPeriodEnded(ctx sdk.Context, proposalID uint64) {
	msgBz, err := json.Marshal(SudoAfterProposalVotingPeriodEnded{
		AfterProposalVotingPeriodEnded: strconv.Itoa(int(proposalID)),
	})
	if err != nil {
		return
	}

	h.sendMsgToAll(ctx, msgBz)
}
