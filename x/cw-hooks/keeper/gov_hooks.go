package keeper

import (
	"encoding/json"
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/CosmosContracts/juno/v27/x/cw-hooks/types"
)

type GovHooks struct {
	k Keeper
}

var _ govtypes.GovHooks = GovHooks{}

func (k Keeper) GovHooks() GovHooks {
	return GovHooks{k: k}
}

type Proposal struct {
	ProposalID uint64 `json:"proposal_id"`
	Proposer   string `json:"proposer"`
	Status     uint   `json:"status"`
	SubmitTime string `json:"submit_time"`
	Metadata   string `json:"metadata"`
	Title      string `json:"title"`
	Summary    string `json:"summary"`
}

func NewProposal(prop v1.Proposal) Proposal {
	return Proposal{
		ProposalID: prop.Id,
		Proposer:   prop.Proposer,
		Status:     uint(prop.Status),
		SubmitTime: strconv.Itoa(prop.SubmitTime.Second()),
		Metadata:   prop.GetMetadata(),
		Title:      prop.GetTitle(),
		Summary:    prop.GetSummary(),
	}
}

type Vote struct {
	ProposalID   uint64                   `json:"proposal_id"`
	VoterAddress string                   `json:"voter_address"`
	VoteOption   []*v1.WeightedVoteOption `json:"vote_option"`
}

func NewVote(vote v1.Vote) Vote {
	return Vote{
		ProposalID:   vote.ProposalId,
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

	if err := h.k.ExecuteMessageOnContracts(ctx, types.KeyPrefixGov, msgBz); err != nil {
		fmt.Println("AfterProposalSubmission: ", err)
		return
	}
}

func (h GovHooks) AfterProposalDeposit(ctx sdk.Context, proposalID uint64, _ sdk.AccAddress) {
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

	if err := h.k.ExecuteMessageOnContracts(ctx, types.KeyPrefixGov, msgBz); err != nil {
		fmt.Println("AfterProposalDeposit: ", err)
		return
	}
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

	if err := h.k.ExecuteMessageOnContracts(ctx, types.KeyPrefixGov, msgBz); err != nil {
		fmt.Println("AfterProposalVote: ", err)
		return
	}
}

func (h GovHooks) AfterProposalFailedMinDeposit(_ sdk.Context, _ uint64) {
}

func (h GovHooks) AfterProposalVotingPeriodEnded(ctx sdk.Context, proposalID uint64) {
	msgBz, err := json.Marshal(SudoAfterProposalVotingPeriodEnded{
		AfterProposalVotingPeriodEnded: strconv.Itoa(int(proposalID)),
	})
	if err != nil {
		return
	}

	if err := h.k.ExecuteMessageOnContracts(ctx, types.KeyPrefixGov, msgBz); err != nil {
		fmt.Println("AfterProposalVotingPeriodEnded: ", err)
		return
	}
}
