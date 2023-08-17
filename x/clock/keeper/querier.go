package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v17/x/clock/types"
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

// ContractModules returns contract addresses which are using the clock
func (q Querier) ClockContracts(stdCtx context.Context, _ *types.QueryClockContracts) (*types.QueryClockContractsResponse, error) {
	ctx := sdk.UnwrapSDKContext(stdCtx)

	p := q.keeper.GetParams(ctx)

	return &types.QueryClockContractsResponse{
		ContractAddresses: p.ContractAddresses,
	}, nil
}
