package types

import (
	"strings"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewParams(
	mintDenom string, blocksPerYear uint64,
) Params {
	return Params{
		MintDenom:     mintDenom,
		BlocksPerYear: blocksPerYear,
	}
}

// DefaultParams are the default module parameters for x/mint
func DefaultParams() Params {
	return Params{
		MintDenom:     sdk.DefaultBondDenom,
		BlocksPerYear: uint64(60 * 60 * 8766 / 5), // assuming 5 second block times
	}
}

// Validate validates the x/mint module parameters
func (p Params) Validate() error {
	if err := validateMintDenom(p.MintDenom); err != nil {
		return err
	}
	err := validateBlocksPerYear(p.BlocksPerYear)

	return err
}

func validateMintDenom(i any) error {
	v, ok := i.(string)
	if !ok {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "expected string, got %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "mint denom cannot be blank")
	}
	err := sdk.ValidateDenom(v)
	return err
}

func validateBlocksPerYear(i any) error {
	v, ok := i.(uint64)
	if !ok {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "expected uint64, got %T", i)
	}

	if v == 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "blocks per year must be positive (%d)", v)
	}

	return nil
}
