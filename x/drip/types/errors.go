package types

import (
	errorsmod "cosmossdk.io/errors"
)

// errors
var (
	ErrDripDisabled   = errorsmod.Register(ModuleName, 1, "drip module is disabled by governance")
	ErrDripNotAllowed = errorsmod.Register(ModuleName, 2, "this address is not allowed to use the module, you can request access from governance")
)
