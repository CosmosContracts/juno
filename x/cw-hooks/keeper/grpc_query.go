package keeper

import (
	"context"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CosmosContracts/juno/v30/x/cw-hooks/types"
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

func (q queryServer) Contracts(ctx context.Context, req *types.QueryContractsRequest) (*types.QueryContractsResponse, error) {
	if req == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "request cannot be nil")
	}

	modulePrefix, err := types.ModulePrefixFromModule(req.Module)
	if err != nil {
		return nil, err
	}

	iter, err := q.k.Contracts.Iterate(ctx, collections.NewPrefixedPairRange[[]byte, sdk.AccAddress](modulePrefix.Bytes()))
	if err != nil {
		return nil, err
	}

	contracts, err := iter.Values()
	if err != nil {
		return nil, err
	}

	return &types.QueryContractsResponse{
		Contracts: contracts,
	}, nil
}

func (q queryServer) ContractInfo(ctx context.Context, req *types.QueryContractInfoRequest) (*types.QueryContractInfoResponse, error) {
	if req == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "request cannot be nil")
	}

	modulePrefix, err := types.ModulePrefixFromModule(req.Module)
	if err != nil {
		return nil, err
	}

	accAddr, err := sdk.AccAddressFromBech32(req.ContractAddress)
	if err != nil {
		return nil, err
	}

	contract, err := q.k.Contracts.Get(ctx, types.BuildContractPrimaryKey(modulePrefix, accAddr))
	if err != nil {
		return nil, err
	}

	return &types.QueryContractInfoResponse{
		Contract: contract,
	}, nil
}
