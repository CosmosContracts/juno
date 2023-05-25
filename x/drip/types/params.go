package types

import (
	"fmt"
)

var (
	DefaultEnableDrip       = true
	DefaultAllowedAddresses = []string(nil) // no one allowed
)

// NewParams creates a new Params object
func NewParams(
	enableDrip bool,
	allowedAddresses []string,
) Params {
	return Params{
		EnableDrip:       enableDrip,
		AllowedAddresses: allowedAddresses,
	}
}

// DefaultParams returns default x/drip module parameters.
func DefaultParams() Params {
	return Params{
		EnableDrip:       DefaultEnableDrip,
		AllowedAddresses: DefaultAllowedAddresses,
	}
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateArray(i interface{}) error {
	_, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, address := range i.([]string) {
		if address == "" {
			return fmt.Errorf("address cannot be blank")
		}

		// TODO: Validate address
	}

	return nil
}

func (p Params) Validate() error {
	if err := validateBool(p.EnableDrip); err != nil {
		return err
	}

	err := validateArray(p.AllowedAddresses)
	return err
}
