package globalfee

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v16/x/globalfee/keeper"
	"github.com/CosmosContracts/juno/v16/x/globalfee/types"
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

// MinimumGasPrices return minimum gas prices
func (g GrpcQuerier) MinimumGasPrices(stdCtx context.Context, _ *types.QueryMinimumGasPricesRequest) (*types.QueryMinimumGasPricesResponse, error) {
	var minGasPrices sdk.DecCoins
	ctx := sdk.UnwrapSDKContext(stdCtx)

	p := g.keeper.GetParams(ctx)
	minGasPrices = p.MinimumGasPrices

	return &types.QueryMinimumGasPricesResponse{
		MinimumGasPrices: minGasPrices,
	}, nil
}
