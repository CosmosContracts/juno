package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v28/x/clock/types"
)

// InitGenesis import module genesis
func (k Keeper) InitGenesis(
	ctx sdk.Context,
	data types.GenesisState,
) {
	// Validate init contents
	if err := types.ValidateGenesis(data); err != nil {
		panic(err)
	}

	// Set params
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}
}

// ExportGenesis export module state
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)

	return &types.GenesisState{
		Params: params,
	}
}
