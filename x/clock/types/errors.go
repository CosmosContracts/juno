package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrContractJailed         = errorsmod.Register(ModuleName, 1, "contract is jailed")
	ErrContractNotJailed      = errorsmod.Register(ModuleName, 2, "contract is not jailed")
	ErrContractAlreadyJailed  = errorsmod.Register(ModuleName, 3, "contract is already jailed")
	ErrMaxContractsRegistered = errorsmod.Register(ModuleName, 4, "maximum number of registered clock contracts reached")
)
