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

// ValidateBasic determines whether the information in the message is formatted correctly:
// the authority must be a valid acc-address and the proposed params must pass their own
// validation. Without the params check, governance could store params (zero window, empty
// fee denom, nil decimals, ...) that panic or deterministically fail in the ante/post
// handlers and EndBlock — halting the chain.
func (m *MsgUpdateParams) ValidateBasic() error {
	// validate authority address
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return err
	}

	return m.Params.ValidateBasic()
}
