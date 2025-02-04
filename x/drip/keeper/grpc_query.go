package keeper

import (
	"context"

	"github.com/CosmosContracts/juno/v27/x/drip/types"
)

var _ types.QueryServer = queryServer{}

func NewQueryServerImpl(k Keeper) types.QueryServer {
	return queryServer{k}
}

type queryServer struct {
	k Keeper
}

// Params returns the fees module params
func (q queryServer) Params(
	c context.Context,
	_ *types.QueryParamsRequest,
) (*types.QueryParamsResponse, error) {
	params := q.k.GetParams(c)
	return &types.QueryParamsResponse{Params: params}, nil
}
