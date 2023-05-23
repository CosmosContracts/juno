package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store key
var (
	DefaultEnableDrip       = true
	DefaultAllowedAddresses = []string(nil) // no one allowed

	ParamStoreKeyEnableDrip       = []byte("EnableFeeShare")
	ParamStoreKeyAllowedAddresses = []byte("AllowedAddresses")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

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

func DefaultParams() Params {
	return Params{
		EnableDrip:       DefaultEnableDrip,
		AllowedAddresses: DefaultAllowedAddresses,
	}
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyEnableDrip, &p.EnableDrip, validateBool),
		paramtypes.NewParamSetPair(ParamStoreKeyAllowedAddresses, &p.AllowedAddresses, validateArray),
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
