package types

import (
	"fmt"
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

type ProposalType string

const (
	ProposalTypeAddTrackingPriceHistory               ProposalType = "AddTrackingPriceHistory"
	ProposalTypeAddTrackingPriceHistoryWithAcceptList ProposalType = "AddTrackingPriceHistoryWithAcceptList"
	ProposalTypeRemoveTrackingPriceHistoryProposal    ProposalType = "RemoveTrackingPriceHistoryProposal"
)

func init() {
	govtypes.RegisterProposalType(string(ProposalTypeAddTrackingPriceHistory))
	govtypes.RegisterProposalType(string(ProposalTypeAddTrackingPriceHistoryWithAcceptList))
	govtypes.RegisterProposalTypeCodec(&AddTrackingPriceHistoryProposal{}, "juno/oracle/AddTrackingPriceHistoryProposal")
	govtypes.RegisterProposalTypeCodec(&AddTrackingPriceHistoryWithAcceptListProposal{}, "juno/oracle/AddTrackingPriceHistoryWithAcceptListProposal")
	govtypes.RegisterProposalTypeCodec(&RemoveTrackingPriceHistoryProposal{}, "juno/oracle/RemoveTrackingPriceHistoryProposal")
}

func NewAddTrackingPriceHistoryProposal(
	title string,
	description string,
	list DenomList,
) *AddTrackingPriceHistoryProposal {
	return &AddTrackingPriceHistoryProposal{title, description, list}
}

// ProposalRoute returns the routing key of a parameter change proposal.
func (p AddTrackingPriceHistoryProposal) ProposalRoute() string { return RouterKey }

// GetTitle returns the title of the proposal
func (p *AddTrackingPriceHistoryProposal) GetTitle() string { return p.Title }

// GetDescription returns the human readable description of the proposal
func (p AddTrackingPriceHistoryProposal) GetDescription() string { return p.Description }

// ProposalType returns the type
func (p AddTrackingPriceHistoryProposal) ProposalType() string {
	return string(ProposalTypeAddTrackingPriceHistory)
}

// ValidateBasic validates the proposal
func (p AddTrackingPriceHistoryProposal) ValidateBasic() error {
	if err := validateProposalCommons(p.Title, p.Description); err != nil {
		return err
	}
	if len(p.TrackingList) == 0 {
		return sdkerrors.Wrap(ErrEmpty, "code updates")
	}
	return nil
}

// String implements the Stringer interface.
func (p AddTrackingPriceHistoryProposal) String() string {
	return fmt.Sprintf(`AddTrackingPriceHistoryProposal:
	Title: 			%s
	Description : 	%s
	TrackingList: 	%v
	`, p.Title, p.Description, p.TrackingList)
}

func NewAddTrackingPriceHistoryWithAcceptListProposal(
	title string,
	description string,
	list DenomList,
) *AddTrackingPriceHistoryWithAcceptListProposal {
	return &AddTrackingPriceHistoryWithAcceptListProposal{title, description, list}
}

// ProposalRoute returns the routing key of a parameter change proposal.
func (p AddTrackingPriceHistoryWithAcceptListProposal) ProposalRoute() string { return RouterKey }

// GetTitle returns the title of the proposal
func (p *AddTrackingPriceHistoryWithAcceptListProposal) GetTitle() string { return p.Title }

// GetDescription returns the human readable description of the proposal
func (p AddTrackingPriceHistoryWithAcceptListProposal) GetDescription() string { return p.Description }

// ProposalType returns the type
func (p AddTrackingPriceHistoryWithAcceptListProposal) ProposalType() string {
	return string(ProposalTypeAddTrackingPriceHistoryWithAcceptList)
}

// ValidateBasic validates the proposal
func (p AddTrackingPriceHistoryWithAcceptListProposal) ValidateBasic() error {
	if err := validateProposalCommons(p.Title, p.Description); err != nil {
		return err
	}
	if len(p.TrackingList) == 0 {
		return sdkerrors.Wrap(ErrEmpty, "code updates")
	}
	return nil
}

// String implements the Stringer interface.
func (p AddTrackingPriceHistoryWithAcceptListProposal) String() string {
	return fmt.Sprintf(`AddTrackingPriceHistoryWithAcceptListProposal:
	Title: 			%s
	Description : 	%s
	TrackingList: 	%v
	`, p.Title, p.Description, p.TrackingList)
}

func NewRemoveTrackingPriceHistoryProposal(
	title string,
	description string,
	list DenomList,
) *RemoveTrackingPriceHistoryProposal {
	return &RemoveTrackingPriceHistoryProposal{title, description, list}
}

// ProposalRoute returns the routing key of a parameter change proposal.
func (p RemoveTrackingPriceHistoryProposal) ProposalRoute() string { return RouterKey }

// GetTitle returns the title of the proposal
func (p *RemoveTrackingPriceHistoryProposal) GetTitle() string { return p.Title }

// GetDescription returns the human readable description of the proposal
func (p RemoveTrackingPriceHistoryProposal) GetDescription() string { return p.Description }

// ProposalType returns the type
func (p RemoveTrackingPriceHistoryProposal) ProposalType() string {
	return string(ProposalTypeRemoveTrackingPriceHistoryProposal)
}

// ValidateBasic validates the proposal
func (p RemoveTrackingPriceHistoryProposal) ValidateBasic() error {
	if err := validateProposalCommons(p.Title, p.Description); err != nil {
		return err
	}
	if len(p.RemoveTrackingList) == 0 {
		return sdkerrors.Wrap(ErrEmpty, "code updates")
	}
	return nil
}

// String implements the Stringer interface.
func (p RemoveTrackingPriceHistoryProposal) String() string {
	return fmt.Sprintf(`RemoveTrackingPriceHistoryProposal:
	Title: 					%s
	Description : 			%s
	RemoveTrackingList: 	%v
	`, p.Title, p.Description, p.RemoveTrackingList)
}

func validateProposalCommons(title, description string) error {
	if strings.TrimSpace(title) != title {
		return sdkerrors.Wrap(govtypes.ErrInvalidProposalContent, "proposal title must not start/end with white spaces")
	}
	if len(title) == 0 {
		return sdkerrors.Wrap(govtypes.ErrInvalidProposalContent, "proposal title cannot be blank")
	}
	if len(title) > govtypes.MaxTitleLength {
		return sdkerrors.Wrapf(govtypes.ErrInvalidProposalContent, "proposal title is longer than max length of %d", govtypes.MaxTitleLength)
	}
	if strings.TrimSpace(description) != description {
		return sdkerrors.Wrap(govtypes.ErrInvalidProposalContent, "proposal description must not start/end with white spaces")
	}
	if len(description) == 0 {
		return sdkerrors.Wrap(govtypes.ErrInvalidProposalContent, "proposal description cannot be blank")
	}
	if len(description) > govtypes.MaxDescriptionLength {
		return sdkerrors.Wrapf(govtypes.ErrInvalidProposalContent, "proposal description is longer than max length of %d", govtypes.MaxDescriptionLength)
	}
	return nil
}
