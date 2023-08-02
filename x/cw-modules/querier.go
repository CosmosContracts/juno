package cwmodules

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v16/x/cw-modules/keeper"
	"github.com/CosmosContracts/juno/v16/x/cw-modules/types"
)

var _ types.QueryServer = &GrpcQuerier{}

type GrpcQuerier struct {
	keeper keeper.Keeper
}

func NewGrpcQuerier(k keeper.Keeper) GrpcQuerier {
	return GrpcQuerier{
		keeper: k,
	}
}

// ContractModules return contract addresses which are uploaded modules
func (g GrpcQuerier) ContractModules(stdCtx context.Context, _ *types.QueryContractModules) (*types.QueryContractModulesResponse, error) {
	ctx := sdk.UnwrapSDKContext(stdCtx)

	p := g.keeper.GetParams(ctx)

	return &types.QueryContractModulesResponse{
		ContractAddresses: p.ContractAddresses,
	}, nil
}
