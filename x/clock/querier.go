package clock

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v17/x/clock/keeper"
	"github.com/CosmosContracts/juno/v17/x/clock/types"
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

// ContractModules returns contract addresses which are using the clock
func (g GrpcQuerier) ClockContracts(stdCtx context.Context, _ *types.QueryClockContracts) (*types.QueryClockContractsResponse, error) {
	ctx := sdk.UnwrapSDKContext(stdCtx)

	p := g.keeper.GetParams(ctx)

	return &types.QueryClockContractsResponse{
		ContractAddresses: p.ContractAddresses,
	}, nil
}
