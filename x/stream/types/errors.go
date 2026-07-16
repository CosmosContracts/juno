package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrCircuitBreakerOpen         = errorsmod.Register(ModuleName, 1, "circuit breaker open")
	ErrCircuitBreakerUnknownState = errorsmod.Register(ModuleName, 2, "unknown circuit breaker state")
	ErrConnectionLimit            = errorsmod.Register(ModuleName, 3, "connection limit reached")
	ErrSubscriptionLimit          = errorsmod.Register(ModuleName, 4, "subscription limit reached")
	ErrMissingModuleOrMethod      = errorsmod.Register(ModuleName, 10, "module and method are required")
	ErrStreamUnavailable          = errorsmod.Register(ModuleName, 11, "stream not available")
	ErrDescriptorRequired         = errorsmod.Register(ModuleName, 12, "descriptor is required")
	ErrUnknownParameter           = errorsmod.Register(ModuleName, 13, "unknown parameter")
	ErrMissingPathParameter       = errorsmod.Register(ModuleName, 14, "missing path parameter")
	ErrMissingQueryParameter      = errorsmod.Register(ModuleName, 15, "missing query parameter")
)
