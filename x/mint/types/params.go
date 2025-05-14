package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewParams(
	mintDenom string, blocksPerYear uint64,
) Params {
	return Params{
		MintDenom:     mintDenom,
		BlocksPerYear: blocksPerYear,
	}
}

// default minting module parameters
func DefaultParams() Params {
	return Params{
		MintDenom:     sdk.DefaultBondDenom,
		BlocksPerYear: uint64(60 * 60 * 8766 / 5), // assuming 5 second block times
	}
}

// validate params
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
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return errors.New("mint denom cannot be blank")
	}
	err := sdk.ValidateDenom(v)
	return err
}

func validateBlocksPerYear(i any) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("blocks per year must be positive: %d", v)
	}

	return nil
}
