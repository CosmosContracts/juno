package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v16/x/globalfee/types"
)

var _ types.QueryServer = Querier{}

// Querier defines a wrapper around the x/FeeShare keeper providing gRPC method
// handlers.
type Querier struct {
	Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

// MinimumGasPrices(context.Context, *QueryMinimumGasPricesRequest) (*QueryMinimumGasPricesResponse, error)

func (q Querier) MinimumGasPrices(c context.Context, req *types.QueryMinimumGasPricesRequest) (*types.QueryMinimumGasPricesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	minGasPrices := q.Keeper.GetParams(ctx).MinimumGasPrices

	return &types.QueryMinimumGasPricesResponse{
		MinimumGasPrices: minGasPrices,
	}, nil
}
