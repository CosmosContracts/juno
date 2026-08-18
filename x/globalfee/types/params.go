package types

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	errDuplicateDenomination = errors.New("duplicate denomination")
	errUnsortedDenomination  = errors.New("is not sorted")
	errNegativeCoinAmount    = errors.New("amount is negative")
)

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{MinimumGasPrices: sdk.DecCoins(nil)}
}

// Validate performs basic validation.
func (p Params) Validate() error {
	return ValidateMinimumGasPrices(p.MinimumGasPrices)
}

// ValidateMinimumGasPrices requires non-negative fees.
func ValidateMinimumGasPrices(i any) error {
	v, ok := i.(sdk.DecCoins)
	if !ok {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidType, "type: %T, expected sdk.DecCoins", i)
	}

	dec := DecCoins(v)
	return dec.Validate()
}

type DecCoins sdk.DecCoins

// Validate checks that the DecCoins are sorted, have nonnegtive amount, with a valid and unique
// denomination (i.e no duplicates). Otherwise, it returns an error.
func (coins DecCoins) Validate() error {
	if len(coins) == 0 {
		return nil
	}

	lowDenom := ""
	seenDenoms := make(map[string]bool)

	for i, coin := range coins {
		if seenDenoms[coin.Denom] {
			return fmt.Errorf("%w %s", errDuplicateDenomination, coin.Denom)
		}
		if err := sdk.ValidateDenom(coin.Denom); err != nil {
			return err
		}
		// skip the denom order check for the first denom in the coins list
		if i != 0 && coin.Denom <= lowDenom {
			return fmt.Errorf("denomination %s %w", coin.Denom, errUnsortedDenomination)
		}
		if coin.IsNegative() {
			return fmt.Errorf("coin %s %w", coin.Amount, errNegativeCoinAmount)
		}

		// we compare each coin against the last denom
		lowDenom = coin.Denom
		seenDenoms[coin.Denom] = true
	}

	return nil
}
