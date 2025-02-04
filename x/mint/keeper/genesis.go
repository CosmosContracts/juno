package keeper

import (
	"context"

	"github.com/CosmosContracts/juno/v27/x/mint/types"
)

// InitGenesis new mint genesis
func (keeper Keeper) InitGenesis(ctx context.Context, ak types.AccountKeeper, data *types.GenesisState) {
	keeper.SetMinter(ctx, data.Minter)

	if err := keeper.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	ak.GetModuleAccount(ctx, types.ModuleName)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func (keeper Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	minter, err := keeper.GetMinter(ctx)
	if err != nil {
		panic(err)
	}
	params, err := keeper.GetParams(ctx)
	if err != nil {
		panic(err)
	}
	return types.NewGenesisState(minter, params)
}
