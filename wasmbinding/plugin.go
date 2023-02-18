package wasmbinding

import (
	wasmKeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmosContracts/juno/v13/wasmbinding/gov"
)

// type FeeShareKeeperExpected interface {
// 	feeshareType.KeeperWriterExpected
// }

type GovKeeperExpected interface {
	gov.KeeperReaderExpected
}

// BuildWasmOptions returns x/wasmd module options to support WASM bindings functionality.
// func BuildWasmOptions(fKeeper FeeShareKeeperExpected, govKeeper GovKeeperExpected) []wasmKeeper.Option {
func BuildWasmOptions(govKeeper GovKeeperExpected) []wasmKeeper.Option {
	return []wasmKeeper.Option{
		// Future: FeeShare bindings
		wasmKeeper.WithQueryPlugins(BuildWasmQueryPlugin(govKeeper)),
	}
}

// BuildWasmQueryPlugin returns the Wasm custom querier plugin.
func BuildWasmQueryPlugin(govKeeper GovKeeperExpected) *wasmKeeper.QueryPlugins {
	return &wasmKeeper.QueryPlugins{
		Custom: NewQueryDispatcher(
			gov.NewQueryHandler(govKeeper),
		).DispatchQuery,
	}
}
