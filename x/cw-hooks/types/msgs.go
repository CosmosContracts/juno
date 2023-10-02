package types

import (
	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// == MsgUpdateParams ==
const TypeMsgUpdateParams = "update_clock_params"

var _ sdk.Msg = &MsgUpdateParams{}

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

// == TypeMsgUnregisterGovernance ==
const TypeMsgUnregisterGovernance = "unregister_governance"

var _ sdk.Msg = &MsgUnregisterGovernance{}

func NewMsgUnregisterGovernance(
	sender sdk.Address,
	contract sdk.Address,
) *MsgUnregisterGovernance {
	return &MsgUnregisterGovernance{
		ContractAddress: contract.String(),
		RegisterAddress: sender.String(),
	}
}

// Route returns the name of the module
func (msg MsgUnregisterGovernance) Route() string { return RouterKey }

// Type returns the the action
func (msg MsgUnregisterGovernance) Type() string { return TypeMsgUnregisterGovernance }

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgUnregisterGovernance) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (msg *MsgUnregisterGovernance) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.RegisterAddress)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (msg *MsgUnregisterGovernance) ValidateBasic() error {
	return nil
}

// == TypeMsgUnregisterStaking ==
const TypeMsgUnregisterStaking = "unregister_staking"

var _ sdk.Msg = &MsgUnregisterStaking{}

func NewMsgUnregisterStaking(
	sender sdk.Address,
	contract sdk.Address,
) *MsgUnregisterStaking {
	return &MsgUnregisterStaking{
		ContractAddress: contract.String(),
		RegisterAddress: sender.String(),
	}
}

// Route returns the name of the module
func (msg MsgUnregisterStaking) Route() string { return RouterKey }

// Type returns the the action
func (msg MsgUnregisterStaking) Type() string { return TypeMsgUnregisterStaking }

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgUnregisterStaking) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (msg *MsgUnregisterStaking) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.RegisterAddress)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (msg *MsgUnregisterStaking) ValidateBasic() error {
	return nil
}
