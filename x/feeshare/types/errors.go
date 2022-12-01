package types

import (
	sdkerrrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// errors
var (
	ErrInternalFeeShare              = sdkerrrors.Register(ModuleName, 2, "internal feeshare error")
	ErrFeeShareDisabled              = sdkerrrors.Register(ModuleName, 3, "feeshare module is disabled by governance")
	ErrFeeShareAlreadyRegistered     = sdkerrrors.Register(ModuleName, 4, "feeshare already exists for given contract")
	ErrFeeShareNoContractDeployed    = sdkerrrors.Register(ModuleName, 5, "no contract deployed")
	ErrFeeShareContractNotRegistered = sdkerrrors.Register(ModuleName, 6, "no feeshare registered for contract")
	ErrFeeShareDeployerIsNotEOA      = sdkerrrors.Register(ModuleName, 7, "no feeshare registered for contract")
)
