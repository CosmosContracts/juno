package types

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

type VoteRequest struct {
	// ProposalID is the unique ID of the proposal.
	ProposalID uint64 `json:"proposal_id"`
	// Voter is the bech32 encoded account address of the voter.
	Voter string `json:"voter"`
}

type (
	VoteResponse struct {
		// Vote defines a vote on a governance proposal.
		Vote Vote `json:"vote,omitempty"`
	}

	Vote struct {
		// ProposalId is the proposal identifier.
		ProposalId uint64 `json:"proposal_id"`
		// Voter is the bech32 encoded account address of the voter.
		Voter string `json:"voter"`
		// Option is the voting option from the enum.
		Options []WeightedVoteOption `json:"options"`
	}

	WeightedVoteOption struct {
		Option VoteOption `json:"option"`
		Weight string     `json:"weight"`
	}
)

type VoteOption int

const (
	Yes VoteOption = iota
	No
	Abstain
	NoWithVeto
)

var fromVoteOption = map[VoteOption]string{
	Yes:        "yes",
	No:         "no",
	Abstain:    "abstain",
	NoWithVeto: "no_with_veto",
}

var toVoteOption = map[string]VoteOption{
	"yes":          Yes,
	"no":           No,
	"abstain":      Abstain,
	"no_with_veto": NoWithVeto,
}

func (v VoteOption) String() string {
	return fromVoteOption[v]
}

func (v VoteOption) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

func (s *VoteOption) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}

	voteOption, ok := toVoteOption[j]
	if !ok {
		return fmt.Errorf("invalid vote option '%v'", j)
	}
	*s = voteOption
	return nil
}

// Validate performs request fields validation.
func (r VoteRequest) Validate() error {
	if r.ProposalID == 0 {
		return fmt.Errorf("proposal_id: must specify a proposal ID to query")
	}

	if _, err := sdk.AccAddressFromBech32(r.Voter); err != nil {
		return fmt.Errorf("voter: parsing: %w", err)
	}

	return nil
}

// MustGetVoter returns the voter as sdk.AccAddress.
// CONTRACT: panics in case of an error (should not happen since we validate the request).
func (r VoteRequest) MustGetVoter() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(r.Voter)
	if err != nil {
		// Should not happen since we validate the request before this call
		panic(fmt.Errorf("wasm bindings: voteRequest request: parsing voter: %w", err))
	}

	return addr
}

func NewVoteResponse(vote govTypes.Vote) VoteResponse {
	resp := VoteResponse{
		Vote: Vote{
			ProposalId: vote.ProposalId,
			Voter:      vote.Voter,
			Options:    make([]WeightedVoteOption, 0, len(vote.Options)),
		},
	}

	for _, option := range vote.Options {
		resp.Vote.Options = append(resp.Vote.Options, NewWeightedVoteOption(option))
	}

	return resp
}

func NewWeightedVoteOption(voteOption govTypes.WeightedVoteOption) WeightedVoteOption {
	var option VoteOption

	switch voteOption.Option {
	case govTypes.OptionYes:
		option = Yes
	case govTypes.OptionNo:
		option = No
	case govTypes.OptionNoWithVeto:
		option = NoWithVeto
	case govTypes.OptionAbstain:
		option = Abstain
	}

	return WeightedVoteOption{
		Option: option,
		Weight: voteOption.Weight.String(),
	}
}
