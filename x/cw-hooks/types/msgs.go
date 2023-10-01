package types

import (
	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// == MsgUpdateParams ==
const TypeMsgUpdateParams = "update_clock_params"

var _ sdk.Msg = &MsgUpdateParams{}

// NewMsgUpdateParams creates new instance of MsgUpdateParams
func NewMsgUpdateParams(
	sender sdk.Address,
) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: sender.String(),
		Params:    Params{},
	}
}

// Route returns the name of the module
func (msg MsgUpdateParams) Route() string { return RouterKey }

// Type returns the the action
func (msg MsgUpdateParams) Type() string { return TypeMsgUpdateParams }

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (msg *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errors.Wrap(err, "invalid authority address")
	}

	return msg.Params.Validate()
}

// == MsgRegisterStaking ==
const TypeMsgRegisterStaking = "register_staking"

var _ sdk.Msg = &MsgRegisterStaking{}

// NewMsgUpdateParams creates new instance of MsgUpdateParams
func NewMsgRegisterStaking(
	sender sdk.Address,
	contract sdk.Address,
) *MsgRegisterStaking {
	return &MsgRegisterStaking{
		RegisterAddress: sender.String(),
		ContractAddress: contract.String(),
	}
}

// Route returns the name of the module
func (msg MsgRegisterStaking) Route() string { return RouterKey }

// Type returns the the action
func (msg MsgRegisterStaking) Type() string { return TypeMsgRegisterStaking }

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgRegisterStaking) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (msg *MsgRegisterStaking) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.RegisterAddress)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (msg *MsgRegisterStaking) ValidateBasic() error {
	return nil
}

// == TypeMsgRegisterGovernance ==
const TypeMsgRegisterGovernance = "register_governance"

var _ sdk.Msg = &MsgRegisterGovernance{}

// NewMsgUpdateParams creates new instance of MsgUpdateParams
func NewMsgRegisterGovernance(
	sender sdk.Address,
	contract sdk.Address,
) *MsgRegisterGovernance {
	return &MsgRegisterGovernance{
		ContractAddress: contract.String(),
		RegisterAddress: sender.String(),
	}
}

// Route returns the name of the module
func (msg MsgRegisterGovernance) Route() string { return RouterKey }

// Type returns the the action
func (msg MsgRegisterGovernance) Type() string { return TypeMsgRegisterGovernance }

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgRegisterGovernance) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (msg *MsgRegisterGovernance) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.RegisterAddress)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (msg *MsgRegisterGovernance) ValidateBasic() error {
	return nil
}
