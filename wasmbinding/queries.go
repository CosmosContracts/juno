package wasmbinding

import (
	oraclekeeper "github.com/CosmosContracts/juno/v12/x/oracle/keeper"
)

type QueryPlugin struct {
	oraclekeeper oraclekeeper.Keeper
}

// NewQueryPlugin returns a reference to a new QueryPlugin.
func NewQueryPlugin(ok oraclekeeper.Keeper) *QueryPlugin {
	return &QueryPlugin{
		oraclekeeper: ok,
	}
}
