package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v18/x/cw-hooks/types"
)

var _ types.QueryServer = &Querier{}

type Querier struct {
	keeper Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{
		keeper: k,
	}
}

// Params returns the total set of clock parameters.
func (q Querier) Params(stdCtx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(stdCtx)

	p := q.keeper.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: &p,
	}, nil
}

func (q Querier) StakingContracts(stdCtx context.Context, _ *types.QueryStakingContractsRequest) (*types.QueryStakingContractsResponse, error) {
	ctx := sdk.UnwrapSDKContext(stdCtx)

	return &types.QueryStakingContractsResponse{
		Contracts: q.keeper.GetAllContractsBech32(ctx, types.KeyPrefixStaking),
	}, nil
}

func (q Querier) GovernanceContracts(stdCtx context.Context, _ *types.QueryGovernanceContractsRequest) (*types.QueryGovernanceContractsResponse, error) {
	ctx := sdk.UnwrapSDKContext(stdCtx)

	return &types.QueryGovernanceContractsResponse{
		Contracts: q.keeper.GetAllContractsBech32(ctx, types.KeyPrefixGov),
	}, nil
}
