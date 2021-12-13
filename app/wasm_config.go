package app

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
)

const (
	// DefaultJunoInstanceCost is initially set the same as in wasmd
	DefaultJunoInstanceCost uint64 = 60_000
	// DefaultJunoCompileCost set to a large number for testing
	DefaultJunoCompileCost uint64 = 100
)

// JunoGasRegisterConfig is defaults plus a custom compile amount
func JunoGasRegisterConfig() wasmkeeper.WasmGasRegisterConfig {
	gasConfig := wasmkeeper.DefaultGasRegisterConfig()
	gasConfig.InstanceCost = DefaultJunoInstanceCost
	gasConfig.CompileCost = DefaultJunoCompileCost

	return gasConfig
}

func NewJunoWasmGasRegister() wasmkeeper.WasmGasRegister {
	return wasmkeeper.NewWasmGasRegister(JunoGasRegisterConfig())
}
