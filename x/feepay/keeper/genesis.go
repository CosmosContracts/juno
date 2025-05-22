package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v30/x/feepay/types"
)

// InitGenesis import module genesis
func (k Keeper) InitGenesis(
	ctx sdk.Context,
	data types.GenesisState,
) {
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	for _, feepay := range data.FeePayContracts {
		// TODO: future, add all wallet interactions for exports?
		k.SetFeePayContract(ctx, feepay)
	}
}

// ExportGenesis export module state
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)
	contracts := k.GetAllContracts(ctx)

	return &types.GenesisState{
		Params:          params,
		FeePayContracts: contracts,
	}
}
