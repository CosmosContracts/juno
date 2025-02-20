package types

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValidateBasic runs stateless checks on the message
func (msg MsgRegisterFeeShare) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DeployerAddress); err != nil {
		return errorsmod.Wrapf(err, "invalid deployer address %s", msg.DeployerAddress)
	}

	if _, err := sdk.AccAddressFromBech32(msg.ContractAddress); err != nil {
		return errorsmod.Wrapf(err, "invalid contract address %s", msg.ContractAddress)
	}

	if msg.WithdrawerAddress != "" {
		if _, err := sdk.AccAddressFromBech32(msg.WithdrawerAddress); err != nil {
			return errorsmod.Wrapf(err, "invalid withdraw address %s", msg.WithdrawerAddress)
		}
	}

	return nil
}

// ValidateBasic runs stateless checks on the message
func (msg MsgCancelFeeShare) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DeployerAddress); err != nil {
		return errorsmod.Wrapf(err, "invalid deployer address %s", msg.DeployerAddress)
	}

	if _, err := sdk.AccAddressFromBech32(msg.ContractAddress); err != nil {
		return errorsmod.Wrapf(err, "invalid deployer address %s", msg.DeployerAddress)
	}

	return nil
}

// ValidateBasic runs stateless checks on the message
func (msg MsgUpdateFeeShare) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DeployerAddress); err != nil {
		return errorsmod.Wrapf(err, "invalid deployer address %s", msg.DeployerAddress)
	}

	if _, err := sdk.AccAddressFromBech32(msg.ContractAddress); err != nil {
		return errorsmod.Wrapf(err, "invalid contract address %s", msg.ContractAddress)
	}

	if _, err := sdk.AccAddressFromBech32(msg.WithdrawerAddress); err != nil {
		return errorsmod.Wrapf(err, "invalid withdraw address %s", msg.WithdrawerAddress)
	}

	return nil
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	return m.Params.Validate()
}
