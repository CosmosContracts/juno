package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	globalerrors "github.com/CosmosContracts/juno/v28/app/helpers"
	"github.com/CosmosContracts/juno/v28/x/clock/types"
)

var _ types.QueryServer = queryServer{}

func NewQueryServerImpl(k Keeper) types.QueryServer {
	return queryServer{k}
}

type queryServer struct {
	k Keeper
}

// ContractModules returns contract addresses which are using the clock
func (q queryServer) ClockContracts(ctx context.Context, req *types.QueryClockContractsRequest) (*types.QueryClockContractsResponse, error) {
	contracts, err := q.k.GetPaginatedContracts(ctx, req.Pagination)
	if err != nil {
		return nil, err
	}

	return contracts, nil
}

// ClockContract returns the clock contract information
func (q queryServer) ClockContract(ctx context.Context, req *types.QueryClockContractRequest) (*types.QueryClockContractResponse, error) {
	// Ensure the contract address is valid
	if _, err := sdk.AccAddressFromBech32(req.ContractAddress); err != nil {
		return nil, globalerrors.ErrInvalidAddress
	}

	contract, err := q.k.GetClockContract(ctx, req.ContractAddress)
	if err != nil {
		return nil, err
	}

	return &types.QueryClockContractResponse{
		ClockContract: *contract,
	}, nil
}

// Params returns the total set of clock parameters.
func (q queryServer) Params(ctx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	p := q.k.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: p,
	}, nil
}
