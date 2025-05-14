package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.Validate() != nil {
		return fmt.Errorf("invalid denom creation fee: %+v", i)
	}

	return nil
}
