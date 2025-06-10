package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidStreamType = errorsmod.Register(ModuleName, 0, "invalid stream type")
	ErrStreamNotFound    = errorsmod.Register(ModuleName, 1, "stream not found")
	ErrNoQueryContext    = errorsmod.Register(ModuleName, 2, "no query context")
)
