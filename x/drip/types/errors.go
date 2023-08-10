package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrDripDisabled   = errorsmod.Register(ModuleName, 1, "drip module is disabled by governance")
	ErrDripNotAllowed = errorsmod.Register(ModuleName, 2, "this address is not allowed to use the module, you can request access from governance")
	ErrEmpty          = errorsmod.Register(ModuleName, 3, "empty")
	ErrDuplicate      = errorsmod.Register(ModuleName, 4, "duplicate")
	ErrBlank          = errorsmod.Register(ModuleName, 5, "address cannot be blank")
)
