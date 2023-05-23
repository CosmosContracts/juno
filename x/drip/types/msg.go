package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgDistributeTokens{}
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

	// TODO: Whitelist trough governance
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
