package wasmbinding

import (
	tokenfactorykeeper "github.com/CosmWasm/token-factory/x/tokenfactory/keeper"
	oraclekeeper "github.com/CosmosContracts/juno/v12/x/oracle/keeper"

	tokenfactorybindings "github.com/CosmWasm/token-factory/x/tokenfactory/bindings"
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

func RegisterCustomPlugins(bank *bankkeeper.BaseKeeper, oracle oraclekeeper.Keeper, tokenFactory *tokenfactorykeeper.Keeper) []wasmkeeper.Option {
	wasmQueryPlugin := NewQueryPlugin(oracle)

	oracleQueryPluginOpt := wasmkeeper.WithQueryPlugins(&wasmkeeper.QueryPlugins{
		Custom: CustomQuerier(wasmQueryPlugin),
	})

	queryPluginOpt := wasmkeeper.WithQueryPlugins(&wasmkeeper.QueryPlugins{
		Custom: CustomQuerier(wasmQueryPlugin),
	})
	messengerDecoratorOpt := wasmkeeper.WithMessageHandlerDecorator(
		tokenfactorybindings.CustomMessageDecorator(bank, tokenFactory),
	)

	return []wasm.Option{
		queryPluginOpt,
		oracleQueryPluginOpt,
		messengerDecoratorOpt,
	}
}
