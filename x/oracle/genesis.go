package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v12/x/oracle/keeper"
	"github.com/CosmosContracts/juno/v12/x/oracle/types"
)

// InitGenesis initializes the x/oracle module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, genState types.GenesisState) {
	keeper.InitGenesis(ctx, genState)
}

// ExportGenesis returns the x/oracle module's exported genesis.
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) *types.GenesisState {
	return keeper.ExportGenesis(ctx)
}
