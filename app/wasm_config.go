package app

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

const (
	// DefaultJunoInstanceCost is initially set the same as in wasmd
	DefaultJunoInstanceCost uint64 = 60_000
	// DefaultJunoCompileCost set to a large number for testing
	DefaultJunoCompileCost uint64 = 3
)

// JunoGasRegisterConfig is defaults plus a custom compile amount
func JunoGasRegisterConfig() wasmtypes.WasmGasRegisterConfig {
	gasConfig := wasmtypes.DefaultGasRegisterConfig()
	gasConfig.InstanceCost = DefaultJunoInstanceCost
	gasConfig.CompileCost = DefaultJunoCompileCost

	return gasConfig
}

func NewJunoWasmGasRegister() wasmtypes.WasmGasRegister {
	return wasmtypes.NewWasmGasRegister(JunoGasRegisterConfig())
}
