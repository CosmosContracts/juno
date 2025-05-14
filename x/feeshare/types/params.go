package types

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
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
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateShares(i any) error {
	v, ok := i.(sdkmath.LegacyDec)

	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return errors.New("invalid parameter: nil")
	}

	if v.IsNegative() {
		return fmt.Errorf("value cannot be negative: %T", i)
	}

	if v.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("value cannot be greater than 1: %T", i)
	}

	return nil
}

func validateArray(i any) error {
	_, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, denom := range i.([]string) {
		if denom == "" {
			return errors.New("denom cannot be blank")
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
