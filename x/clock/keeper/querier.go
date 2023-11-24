package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v18/x/clock/types"
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

	contracts := q.keeper.GetAllContracts(ctx, false)

	return &types.QueryClockContractsResponse{
		ContractAddresses: contracts,
	}, nil
}

// JailedClockContracts returns contract addresses which have been jailed for erroring
func (q Querier) JailedClockContracts(stdCtx context.Context, _ *types.QueryJailedClockContracts) (*types.QueryJailedClockContractsResponse, error) {
	ctx := sdk.UnwrapSDKContext(stdCtx)

	contracts := q.keeper.GetAllContracts(ctx, true)

	return &types.QueryJailedClockContractsResponse{
		JailedContractAddresses: contracts,
	}, nil
}

// ClockContract returns the clock contract information
func (q Querier) ClockContract(stdCtx context.Context, req *types.QueryClockContract) (*types.QueryClockContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(stdCtx)

	// Ensure the contract address is valid
	if _, err := sdk.AccAddressFromBech32(req.ContractAddress); err != nil {
		return nil, types.ErrInvalidAddress
	}

	// Check if the contract is jailed or unjailed
	isUnjailed := q.keeper.IsClockContract(ctx, req.ContractAddress, false)
	isJailed := q.keeper.IsClockContract(ctx, req.ContractAddress, true)

	// Return the registered contract or an error if it doesn't exist
	if isUnjailed || isJailed {
		return &types.QueryClockContractResponse{
			ContractAddress: req.ContractAddress,
			IsJailed:        isJailed,
		}, nil
	}

	return nil, types.ErrContractNotRegistered
}

// Params returns the total set of clock parameters.
func (q Querier) Params(stdCtx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(stdCtx)

	p := q.keeper.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: &p,
	}, nil
}
