package types

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewParams creates new, configurable params for the tokenfactory module.
func NewParams(denomCreationFee sdk.Coins) Params {
	return Params{
		DenomCreationFee: denomCreationFee,
	}
}

// DefaultParams are the tokenfactory default module parameters.
func DefaultParams() Params {
	return Params{
		DenomCreationFee:        sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10_000_000)),
		DenomCreationGasConsume: 2_000_000,
	}
}

// Validate the tokenfactory module parameters.
func (p Params) Validate() error {
	err := validateDenomCreationFee(p.DenomCreationFee)

	return err
}

func validateDenomCreationFee(i any) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "expected sdk.Coins, got %T", i)
	}

	if err := v.Validate(); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid denom creation fee: %+v", err)
	}

	return nil
}
