package keeper

import (
	"context"

	"github.com/CosmosContracts/juno/v29/x/mint/types"
)

var _ types.QueryServer = queryServer{}

func NewQueryServerImpl(k Keeper) types.QueryServer {
	return queryServer{k}
}

type queryServer struct {
	k Keeper
}

// Params returns params of the mint module.
func (q queryServer) Params(ctx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params, err := q.k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryParamsResponse{Params: params}, nil
}

// Inflation returns minter.Inflation of the mint module.
func (q queryServer) Inflation(ctx context.Context, _ *types.QueryInflationRequest) (*types.QueryInflationResponse, error) {
	minter, err := q.k.GetMinter(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryInflationResponse{Inflation: minter.Inflation}, nil
}

// AnnualProvisions returns minter.AnnualProvisions of the mint module.
func (q queryServer) AnnualProvisions(ctx context.Context, _ *types.QueryAnnualProvisionsRequest) (*types.QueryAnnualProvisionsResponse, error) {
	minter, err := q.k.GetMinter(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryAnnualProvisionsResponse{AnnualProvisions: minter.AnnualProvisions}, nil
}

// Target supply returns minter.TargetSupply of the mint module.
func (q queryServer) TargetSupply(ctx context.Context, _ *types.QueryTargetSupplyRequest) (*types.QueryTargetSupplyResponse, error) {
	minter, err := q.k.GetMinter(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryTargetSupplyResponse{TargetSupply: minter.TargetSupply}, nil
}
