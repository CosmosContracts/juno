package feepay

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v19/x/feepay/keeper"
	"github.com/CosmosContracts/juno/v19/x/feepay/types"
)

// InitGenesis import module genesis
func InitGenesis(
	ctx sdk.Context,
	k keeper.Keeper,
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
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	params := k.GetParams(ctx)
	contracts := k.GetAllContracts(ctx)

	return &types.GenesisState{
		Params:          params,
		FeePayContracts: contracts,
	}
}
