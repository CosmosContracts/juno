package types

import (
	sdkerrrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// errors
var (
	ErrInternalRevenue              = sdkerrrors.Register(ModuleName, 2, "internal revenue error")
	ErrRevenueDisabled              = sdkerrrors.Register(ModuleName, 3, "revenue module is disabled by governance")
	ErrRevenueAlreadyRegistered     = sdkerrrors.Register(ModuleName, 4, "revenue already exists for given contract")
	ErrRevenueNoContractDeployed    = sdkerrrors.Register(ModuleName, 5, "no contract deployed")
	ErrRevenueContractNotRegistered = sdkerrrors.Register(ModuleName, 6, "no revenue registered for contract")
	ErrRevenueDeployerIsNotEOA      = sdkerrrors.Register(ModuleName, 7, "no revenue registered for contract")
)
