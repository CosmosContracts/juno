package types

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ValidateBasic does a sanity check on the provided msg.
func (MsgCreateDenom) ValidateBasic() error {
	return nil
}

// ValidateBasic does a sanity check on the provided msg.
func (m MsgMint) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	if m.MintToAddress != "" {
		_, err = sdk.AccAddressFromBech32(m.MintToAddress)
		if err != nil {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid mint to address (%s)", err)
		}
	}

	if !m.Amount.IsValid() || m.Amount.Amount.Equal(sdkmath.ZeroInt()) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, m.Amount.String())
	}

	return nil
}

// ValidateBasic does a sanity check on the provided msg.
func (m MsgBurn) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	if !m.Amount.IsValid() || m.Amount.Amount.Equal(sdkmath.ZeroInt()) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, m.Amount.String())
	}

	if m.BurnFromAddress != "" {
		_, err = sdk.AccAddressFromBech32(m.BurnFromAddress)
		if err != nil {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid burn from address (%s)", err)
		}
	}

	return nil
}

// ValidateBasic does a sanity check on the provided msg.
func (m MsgForceTransfer) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(m.TransferFromAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid from address (%s)", err)
	}
	_, err = sdk.AccAddressFromBech32(m.TransferToAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid to address (%s)", err)
	}

	if !m.Amount.IsValid() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, m.Amount.String())
	}

	return nil
}

// ValidateBasic does a sanity check on the provided msg.
func (m MsgChangeAdmin) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(m.NewAdmin)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid address (%s)", err)
	}

	_, _, err = DeconstructDenom(m.Denom)
	if err != nil {
		return err
	}

	return nil
}

// ValidateBasic does a sanity check on the provided msg.
func (m MsgSetDenomMetadata) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	err = m.Metadata.Validate()
	if err != nil {
		return err
	}

	_, _, err = DeconstructDenom(m.Metadata.Base)
	if err != nil {
		return err
	}

	return nil
}

// ValidateBasic does a sanity check on the provided msg.
func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	return m.Params.Validate()
}
