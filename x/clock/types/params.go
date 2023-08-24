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
		ContractGasLimit:  1_000_000_000, // 1 billion
	}
}

// NewParams creates a new Params object
func NewParams(
	contracts []string,
	contractGasLimit uint64,
) Params {
	return Params{
		ContractAddresses: contracts,
		ContractGasLimit:  contractGasLimit,
	}
}

// Validate performs basic validation.
func (p Params) Validate() error {
	minimumGas := uint64(100_000)
	if p.ContractGasLimit < minimumGas {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"invalid contract gas limit: %d. Must be above %d", p.ContractGasLimit, minimumGas,
		)
	}

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
