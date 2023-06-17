package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgDistributeTokens{}
	_ sdk.Msg = &MsgUpdateParams{}
)

const (
	TypeMsgDistributeTokens = "distribute_tokens"
)

// NewMsgDistributeTokens creates new instance of MsgDistributeTokens
func NewMsgDistributeTokens(
	amount sdk.Coins,
	sender sdk.Address,
) *MsgDistributeTokens {
	return &MsgDistributeTokens{
		SenderAddress: sender.String(),
		Amount:        amount,
	}
}

// Route returns the name of the module
func (msg MsgDistributeTokens) Route() string { return RouterKey }

// Type returns the the action
func (msg MsgDistributeTokens) Type() string { return TypeMsgDistributeTokens }

// ValidateBasic runs stateless checks on the message
func (msg MsgDistributeTokens) ValidateBasic() error {
	// Validation logic is inside msg_server
	return nil
}

// GetSignBytes encodes the message for signing
func (msg *MsgDistributeTokens) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgDistributeTokens) GetSigners() []sdk.AccAddress {
	from, _ := sdk.AccAddressFromBech32(msg.SenderAddress)
	return []sdk.AccAddress{from}
}

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (msg *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	if err := msg.Params.Validate(); err != nil {
		return err
	}

	return nil
}
