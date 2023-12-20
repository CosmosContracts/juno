package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrContractJailed        = errorsmod.Register(ModuleName, 6, "contract is jailed")
	ErrContractNotJailed     = errorsmod.Register(ModuleName, 7, "contract is not jailed")
	ErrContractAlreadyJailed = errorsmod.Register(ModuleName, 8, "contract is already jailed")
)
