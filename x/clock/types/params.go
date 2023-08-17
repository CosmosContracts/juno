package types

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		ContractAddresses: []string(nil),
	}
}

// Validate performs basic validation.
func (p Params) Validate() error {
	// confirm each contract address is a valid length
	for _, addr := range p.ContractAddresses {
		_, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			return errorsmod.Wrapf(
				sdkerrors.ErrInvalidAddress,
				"invalid contract address: %s", err.Error(),
			)
		}
	}

	return nil
}
