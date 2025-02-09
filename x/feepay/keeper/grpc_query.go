package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	globalerrors "github.com/CosmosContracts/juno/v28/app/helpers"
	"github.com/CosmosContracts/juno/v28/x/feepay/types"
)

var _ types.QueryServer = queryServer{}

func NewQueryServerImpl(k Keeper) types.QueryServer {
	return queryServer{k}
}

type queryServer struct {
	k Keeper
}

// FeePayContract implements types.QueryServer.
func (q queryServer) FeePayContract(ctx context.Context, req *types.QueryFeePayContractRequest) (*types.QueryFeePayContractResponse, error) {
	// Check if contract address are valid
	if _, err := sdk.AccAddressFromBech32(req.ContractAddress); err != nil {
		return nil, globalerrors.ErrInvalidAddress
	}

	contract, err := q.k.GetContract(ctx, req.ContractAddress)
	if err != nil {
		return nil, err
	}

	return &types.QueryFeePayContractResponse{
		FeePayContract: *contract,
	}, nil
}

// FeePayContracts implements types.QueryServer.
func (q queryServer) FeePayContracts(ctx context.Context, req *types.QueryFeePayContractsRequest) (*types.QueryFeePayContractsResponse, error) {
	res, err := q.k.GetContracts(ctx, req.Pagination)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// FeePayContractUses implements types.QueryServer.
func (q queryServer) FeePayContractUses(ctx context.Context, req *types.QueryFeePayContractUsesRequest) (*types.QueryFeePayContractUsesResponse, error) {
	// Check if wallet & contract address are valid
	if _, err := sdk.AccAddressFromBech32(req.ContractAddress); err != nil {
		return nil, globalerrors.ErrInvalidAddress
	}

	if _, err := sdk.AccAddressFromBech32(req.WalletAddress); err != nil {
		return nil, globalerrors.ErrInvalidAddress
	}

	// Get contract from KV store
	fpc, err := q.k.GetContract(ctx, req.ContractAddress)
	if err != nil {
		return nil, err
	}

	uses, err := q.k.GetContractUses(ctx, fpc, req.WalletAddress)
	if err != nil {
		return nil, err
	}

	return &types.QueryFeePayContractUsesResponse{
		Uses: uses,
	}, nil
}

// FeePayContractEligible implements types.QueryServer.
func (q queryServer) FeePayWalletIsEligible(ctx context.Context, req *types.QueryFeePayWalletIsEligibleRequest) (*types.QueryFeePayWalletIsEligibleResponse, error) {
	// Check if wallet & contract address are valid
	if _, err := sdk.AccAddressFromBech32(req.ContractAddress); err != nil {
		return nil, globalerrors.ErrInvalidAddress
	}

	if _, err := sdk.AccAddressFromBech32(req.WalletAddress); err != nil {
		return nil, globalerrors.ErrInvalidAddress
	}

	// Get fee pay contract
	fpc, err := q.k.GetContract(ctx, req.ContractAddress)
	if err != nil {
		return nil, err
	}

	// Return if wallet is eligible
	isEligible, err := q.k.IsWalletEligible(ctx, fpc, req.WalletAddress)
	if err != nil {
		return nil, err
	}

	return &types.QueryFeePayWalletIsEligibleResponse{
		Eligible: isEligible,
	}, nil
}

// Params returns the feepay module params
func (q queryServer) Params(
	c context.Context,
	_ *types.QueryParamsRequest,
) (*types.QueryParamsResponse, error) {
	params := q.k.GetParams(c)
	return &types.QueryParamsResponse{Params: params}, nil
}
