package types

import (
	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v19/x/drip/types"
)

// == MsgUploadContentBlob ==
const TypeMsgUploadContentBlob = "upload_content"

var _ sdk.Msg = &MsgUploadContentBlob{}

// NewMsgUploadContentBlob creates new instance of MsgUploadContentBlob
func NewMsgUploadContentBlob(
	sender sdk.Address,
	idKey uint64,
	content []byte,
) *MsgUploadContentBlob {
	return &MsgUploadContentBlob{
		Sender:  sender.String(),
		IdKey:   idKey,
		Content: content,
	}
}

// Route returns the name of the module
func (msg MsgUploadContentBlob) Route() string { return types.RouterKey }

// Type returns the the action
func (msg MsgUploadContentBlob) Type() string { return TypeMsgUploadContentBlob }

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgUploadContentBlob) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for a MsgUploadContentBlob message.
func (msg *MsgUploadContentBlob) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (msg *MsgUploadContentBlob) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return errors.Wrap(err, "invalid sender address")
	}

	// return msg.Params.Validate()
	return nil
}

// == MsgUploadContentBlob ==
const TypeMsgUploadMetadata = "upload_metadata"

var _ sdk.Msg = &MsgUploadMetadata{}

// NewMsgUploadContentBlob creates new instance of MsgUploadContentBlob
func NewMsgUploadMetadata(
	sender sdk.Address,
	title, description, creator string,
	videoIdStart, videoIdEnd uint64,
) *MsgUploadMetadata {
	return &MsgUploadMetadata{
		Sender:      sender.String(),
		Title:       title,
		Description: description,

		IdStart: videoIdStart,
		IdEnd:   videoIdEnd,
	}
}

// Route returns the name of the module
func (msg MsgUploadMetadata) Route() string { return types.RouterKey }

// Type returns the the action
func (msg MsgUploadMetadata) Type() string { return TypeMsgUploadMetadata }

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgUploadMetadata) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for a MsgUploadContentBlob message.
func (msg *MsgUploadMetadata) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (msg *MsgUploadMetadata) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return errors.Wrap(err, "invalid sender address")
	}

	// return msg.Params.Validate()
	return nil
}
