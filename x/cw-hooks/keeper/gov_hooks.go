package keeper

import (
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
	proposalId uint64 `json:"proposal_id"`
	proposer   string `json:"proposer"`
	status     uint   `json:"status"`
	submitTime uint64 `json:"submit_time"`
	metadata   string `json:"metadata"`
	title      string `json:"title"`
	summary    string `json:"summary"`
}

func NewProposal(prop v1.Proposal) Proposal {
	return Proposal{
		proposalId: prop.Id,
		proposer:   prop.Proposer,
		status:     uint(prop.Status),
		// submitTime: prop.SubmitTime,
		metadata: prop.GetMetadata(),
		title:    prop.GetTitle(),
		summary:  prop.GetSummary(),
	}
}

type SudoMsgAfterProposalSubmission struct {
	AfterProposalSubmission Proposal `json:"after_proposal_submission"`
}

func (h GovHooks) AfterProposalSubmission(ctx sdk.Context, proposalID uint64) {
	// prop, found := h.k.govKeeper.GetProposal(ctx, proposalID)
	// if !found {
	// 	return
	// }
}

func (h GovHooks) AfterProposalDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress) {
	// h.AfterProposalDeposit(ctx, proposalID, depositorAddr)
}

func (h GovHooks) AfterProposalVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {
	// for i := range h {
	// 	h[i].AfterProposalVote(ctx, proposalID, voterAddr)
	// }
}

func (h GovHooks) AfterProposalFailedMinDeposit(ctx sdk.Context, proposalID uint64) {
	// for i := range h {
	// 	h[i].AfterProposalFailedMinDeposit(ctx, proposalID)
	// }
}

func (h GovHooks) AfterProposalVotingPeriodEnded(ctx sdk.Context, proposalID uint64) {
	// for i := range h {
	// 	h[i].AfterProposalVotingPeriodEnded(ctx, proposalID)
	// }
}
