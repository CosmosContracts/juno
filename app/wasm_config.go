package app

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
)

const (
	DefaultJunoCompileCost uint64 = 4
)

// JunoGasRegisterConfig is defaults plus a custom compile amount
func JunoGasRegisterConfig() wasmkeeper.WasmGasRegisterConfig {
	return wasmkeeper.WasmGasRegisterConfig{
		InstanceCost:               wasmkeeper.DefaultInstanceCost,
		CompileCost:                DefaultJunoCompileCost,
		GasMultiplier:              wasmkeeper.DefaultGasMultiplier,
		EventPerAttributeCost:      wasmkeeper.DefaultPerAttributeCost,
		CustomEventCost:            wasmkeeper.DefaultPerCustomEventCost,
		EventAttributeDataCost:     wasmkeeper.DefaultEventAttributeDataCost,
		EventAttributeDataFreeTier: wasmkeeper.DefaultEventAttributeDataFreeTier,
		ContractMessageDataCost:    wasmkeeper.DefaultContractMessageDataCost,
	}
}

func NewJunoWasmGasRegister() wasmkeeper.WasmGasRegister {
	return wasmkeeper.NewWasmGasRegister(JunoGasRegisterConfig())
}
