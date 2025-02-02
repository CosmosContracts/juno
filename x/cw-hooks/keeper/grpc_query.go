package keeper

import (
	"context"

	"github.com/CosmosContracts/juno/v27/x/cw-hooks/types"
)

var _ types.QueryServer = queryServer{}

func NewQueryServerImpl(k Keeper) types.QueryServer {
	return queryServer{k}
}

type queryServer struct {
	k Keeper
}

// Params returns the total set of clock parameters.
func (q queryServer) Params(ctx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	p := q.k.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: p,
	}, nil
}

func (q queryServer) StakingContracts(ctx context.Context, _ *types.QueryStakingContractsRequest) (*types.QueryStakingContractsResponse, error) {
	return &types.QueryStakingContractsResponse{
		Contracts: q.k.GetAllContractsBech32(ctx, types.KeyPrefixStaking),
	}, nil
}

func (q queryServer) GovernanceContracts(ctx context.Context, _ *types.QueryGovernanceContractsRequest) (*types.QueryGovernanceContractsResponse, error) {
	return &types.QueryGovernanceContractsResponse{
		Contracts: q.k.GetAllContractsBech32(ctx, types.KeyPrefixGov),
	}, nil
}
