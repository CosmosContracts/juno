package feeshare

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v27/x/feeshare/keeper"
	"github.com/CosmosContracts/juno/v27/x/feeshare/types"
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

	for _, share := range data.FeeShare {
		contract := share.GetContractAddr()
		deployer := share.GetDeployerAddr()
		withdrawer := share.GetWithdrawerAddr()

		// Set initial contracts receiving transaction fees
		k.SetFeeShare(ctx, share)
		k.SetDeployerMap(ctx, deployer, contract)

		if len(withdrawer) != 0 {
			k.SetWithdrawerMap(ctx, withdrawer, contract)
		}
	}
}

// ExportGenesis export module state
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:   k.GetParams(ctx),
		FeeShare: k.GetFeeShares(ctx),
	}
}
