package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrContractNotRegistered     = errorsmod.Register(ModuleName, 1, "contract not registered")
	ErrContractAlreadyRegistered = errorsmod.Register(ModuleName, 2, "contract already registered")
	ErrContractRegisterNotAdmin  = errorsmod.Register(ModuleName, 3, "this address is not the contract admin, cannot register")
	ErrContractNotEnoughFunds    = errorsmod.Register(ModuleName, 4, "contract does not have enough funds")
)
