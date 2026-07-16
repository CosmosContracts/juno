package types

import (
	errorsmod "cosmossdk.io/errors"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DefaultMaxContracts is the default cap on the number of contracts registered
// per hook module (staking/gov). Each registered contract is sudo-executed on
// every matching hook under a child gas meter that is NOT charged to the block
// gas meter, so an unbounded registered set is a block-time DoS vector.
// Governance can raise/lower this via MsgUpdateParams.
const DefaultMaxContracts = uint64(100)

// DefaultParams returns default parameters
func DefaultParams() Params {
	return NewParams(250_000, 3, DefaultMaxContracts)
}

// NewParams creates a new Params object
func NewParams(contractGasLimit uint64, contractFailureRemovalThreshold uint64, maxContracts uint64) Params {
	return Params{
		ContractGasLimit:                contractGasLimit,
		ContractFailureRemovalThreshold: contractFailureRemovalThreshold,
		MaxContracts:                    maxContracts,
	}
}

// Validate performs basic validation.
func (p Params) Validate() error {
	if p.MaxContracts == 0 {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"invalid max contracts: %d. Must be greater than 0", p.MaxContracts,
		)
	}

	return nil
}
