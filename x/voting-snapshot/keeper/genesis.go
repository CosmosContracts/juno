package keeper

import (
	"context"

	"github.com/CosmosContracts/juno/v30/x/voting-snapshot/types"
)

func (k Keeper) InitGenesis(ctx context.Context, gs *types.GenesisState) error {
	return k.Params.Set(ctx, gs.Params)
}

func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &types.GenesisState{Params: params}, nil
}
