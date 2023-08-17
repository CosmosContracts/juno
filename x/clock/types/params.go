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

// NewParams creates a new Params object
func NewParams(
	contracts []string,
) Params {
	return Params{
		ContractAddresses: contracts,
	}
}

// Validate performs basic validation.
func (p Params) Validate() error {
	for _, addr := range p.ContractAddresses {
		// Valid address check
		if _, err := sdk.AccAddressFromBech32(addr); err != nil {
			return errorsmod.Wrapf(
				sdkerrors.ErrInvalidAddress,
				"invalid contract address: %s", err.Error(),
			)
		}

		// duplicate address check
		count := 0
		for _, addr2 := range p.ContractAddresses {
			if addr == addr2 {
				count++
			}

			if count > 1 {
				return errorsmod.Wrapf(
					sdkerrors.ErrInvalidAddress,
					"duplicate contract address: %s", addr,
				)
			}
		}
	}

	return nil
}
