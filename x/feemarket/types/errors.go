package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrNoFeeCoins      = errorsmod.Register(ModuleName, 1, "no fee coin provided. Must provide one.")
	ErrTooManyFeeCoins = errorsmod.Register(ModuleName, 2, "too many fee coins provided.  Only one fee coin may be provided")
	ErrResolverNotSet  = errorsmod.Register(ModuleName, 3, "denom resolver interface not set.  Only the feemarket base fee denomination can be used")
)
