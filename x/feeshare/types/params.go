package types

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	DefaultEnableFeeShare  = true
	DefaultDeveloperShares = sdkmath.LegacyNewDecWithPrec(50, 2) // 50%
	DefaultAllowedDenoms   = []string(nil)                       // all allowed
)

// NewParams creates a new Params object
func NewParams(
	enableFeeShare bool,
	developerShares sdkmath.LegacyDec,
	allowedDenoms []string,
) Params {
	return Params{
		EnableFeeShare:  enableFeeShare,
		DeveloperShares: developerShares,
		AllowedDenoms:   allowedDenoms,
	}
}

func DefaultParams() Params {
	return Params{
		EnableFeeShare:  DefaultEnableFeeShare,
		DeveloperShares: DefaultDeveloperShares,
		AllowedDenoms:   DefaultAllowedDenoms,
	}
}

func validateBool(i any) error {
	_, ok := i.(bool)
	if !ok {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "expected bool, got %T", i)
	}

	return nil
}

func validateShares(i any) error {
	v, ok := i.(sdkmath.LegacyDec)

	if !ok {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "expected sdkmath.LegacyDec, got %T", i)
	}

	if v.IsNil() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid parameter: nil")
	}

	if v.IsNegative() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "value cannot be negative (%s)", v.String())
	}

	if v.GT(sdkmath.LegacyOneDec()) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "value cannot be greater than 1 (%s)", v.String())
	}

	return nil
}

func validateArray(i any) error {
	_, ok := i.([]string)
	if !ok {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "expected []string, got %T", i)
	}

	for _, denom := range i.([]string) {
		if denom == "" {
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "denom cannot be blank")
		}
	}

	return nil
}

func (p Params) Validate() error {
	if err := validateBool(p.EnableFeeShare); err != nil {
		return err
	}
	if err := validateShares(p.DeveloperShares); err != nil {
		return err
	}
	err := validateArray(p.AllowedDenoms)
	return err
}
