package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgUpdateParams{}

// NewMsgParams returns a new message to update the x/feemarket module's parameters.
func NewMsgParams(authority string, params Params) MsgUpdateParams {
	return MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}

// GetSigners implements GetSigners for the msg.
func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic determines whether the information in the message is formatted correctly, specifically
// whether the authority is a valid acc-address.
func (m *MsgUpdateParams) ValidateBasic() error {
	// validate authority address
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return err
	}

	return nil
}
