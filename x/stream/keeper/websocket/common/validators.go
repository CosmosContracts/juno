package common

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AddressValidator provides validation and parsing for blockchain addresses
type AddressValidator struct {
	accAddr *sdk.AccAddress
	valAddr *sdk.ValAddress
	error   error
}

// ValidateAccAddress validates and stores an account address
func ValidateAccAddress(address string) *AddressValidator {
	v := &AddressValidator{}
	if address == "" {
		v.error = fmt.Errorf("address cannot be empty")
		return v
	}

	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		v.error = fmt.Errorf("invalid address: %w", err)
		return v
	}
	v.accAddr = &addr
	return v
}

// ValidateValAddress validates and stores a validator address
func ValidateValAddress(address string) *AddressValidator {
	v := &AddressValidator{}
	if address == "" {
		v.error = fmt.Errorf("validator address cannot be empty")
		return v
	}

	addr, err := sdk.ValAddressFromBech32(address)
	if err != nil {
		v.error = fmt.Errorf("invalid validator address: %w", err)
		return v
	}
	v.valAddr = &addr
	return v
}

// Error returns the validation error if any
func (v *AddressValidator) Error() error {
	return v.error
}

// IsValid returns true if the address is valid
func (v *AddressValidator) IsValid() bool {
	return v.error == nil
}

// AccAddress returns the account address if valid
func (v *AddressValidator) AccAddress() sdk.AccAddress {
	if v.accAddr != nil {
		return *v.accAddr
	}
	return nil
}

// ValAddress returns the validator address if valid
func (v *AddressValidator) ValAddress() sdk.ValAddress {
	if v.valAddr != nil {
		return *v.valAddr
	}
	return nil
}

// ValidationFunc creates a validation function for use in ConnectionParams
func (v *AddressValidator) ValidationFunc() func() error {
	return func() error {
		return v.error
	}
}

// CompositeValidator allows combining multiple validators
type CompositeValidator struct {
	validators []func() error
}

// NewCompositeValidator creates a new composite validator
func NewCompositeValidator(validators ...func() error) *CompositeValidator {
	return &CompositeValidator{validators: validators}
}

// Add adds a validator to the composite
func (c *CompositeValidator) Add(validator func() error) *CompositeValidator {
	c.validators = append(c.validators, validator)
	return c
}

// AddAddressValidator adds an address validator to the composite
func (c *CompositeValidator) AddAddressValidator(v *AddressValidator) *CompositeValidator {
	if v != nil {
		c.validators = append(c.validators, v.ValidationFunc())
	}
	return c
}

// ValidationFunc returns the composite validation function
func (c *CompositeValidator) ValidationFunc() func() error {
	return func() error {
		for _, v := range c.validators {
			if err := v(); err != nil {
				return err
			}
		}
		return nil
	}
}
