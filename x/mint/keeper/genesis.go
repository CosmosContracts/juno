package keeper

import (
	"context"

	"github.com/CosmosContracts/juno/v30/x/mint/types"
)

// InitGenesis new mint genesis
func (k Keeper) InitGenesis(ctx context.Context, ak types.AccountKeeper, data *types.GenesisState) {
	if err := k.SetMinter(ctx, data.Minter); err != nil {
		panic(err)
	}

	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	ak.GetModuleAccount(ctx, types.ModuleName)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func (k Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	minter, err := k.GetMinter(ctx)
	if err != nil {
		panic(err)
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}
	return types.NewGenesisState(minter, params)
}
