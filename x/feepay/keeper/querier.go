package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	globalerrors "github.com/CosmosContracts/juno/v23/app/helpers"
	"github.com/CosmosContracts/juno/v23/x/feepay/types"
)

var _ types.QueryServer = Querier{}

// Querier defines a wrapper around the x/feepay keeper providing gRPC method
// handlers.
type Querier struct {
	Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

// FeePayContract implements types.QueryServer.
func (q Querier) FeePayContract(ctx context.Context, req *types.QueryFeePayContract) (*types.QueryFeePayContractResponse, error) {
	// Check if contract address are valid
	if _, err := sdk.AccAddressFromBech32(req.ContractAddress); err != nil {
		return nil, globalerrors.ErrInvalidAddress
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	contract, err := q.Keeper.GetContract(sdkCtx, req.ContractAddress)
	if err != nil {
		return nil, err
	}

	return &types.QueryFeePayContractResponse{
		FeePayContract: contract,
	}, nil
}

// FeePayContracts implements types.QueryServer.
func (q Querier) FeePayContracts(ctx context.Context, req *types.QueryFeePayContracts) (*types.QueryFeePayContractsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	res, err := q.Keeper.GetContracts(sdkCtx, req.Pagination)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// FeePayContractUses implements types.QueryServer.
func (q Querier) FeePayContractUses(ctx context.Context, req *types.QueryFeePayContractUses) (*types.QueryFeePayContractUsesResponse, error) {
	// Check if wallet & contract address are valid
	if _, err := sdk.AccAddressFromBech32(req.ContractAddress); err != nil {
		return nil, globalerrors.ErrInvalidAddress
	}

	if _, err := sdk.AccAddressFromBech32(req.WalletAddress); err != nil {
		return nil, globalerrors.ErrInvalidAddress
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get contract from KV store
	fpc, err := q.Keeper.GetContract(sdkCtx, req.ContractAddress)
	if err != nil {
		return nil, err
	}

	uses, err := q.Keeper.GetContractUses(sdkCtx, fpc, req.WalletAddress)
	if err != nil {
		return nil, err
	}

	return &types.QueryFeePayContractUsesResponse{
		Uses: uses,
	}, nil
}

// FeePayContractEligible implements types.QueryServer.
func (q Querier) FeePayWalletIsEligible(ctx context.Context, req *types.QueryFeePayWalletIsEligible) (*types.QueryFeePayWalletIsEligibleResponse, error) {
	// Check if wallet & contract address are valid
	if _, err := sdk.AccAddressFromBech32(req.ContractAddress); err != nil {
		return nil, globalerrors.ErrInvalidAddress
	}

	if _, err := sdk.AccAddressFromBech32(req.WalletAddress); err != nil {
		return nil, globalerrors.ErrInvalidAddress
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get fee pay contract
	fpc, err := q.Keeper.GetContract(sdkCtx, req.ContractAddress)
	if err != nil {
		return nil, err
	}

	// Return if wallet is eligible
	isEligible, err := q.Keeper.IsWalletEligible(sdkCtx, fpc, req.WalletAddress)
	if err != nil {
		return nil, err
	}

	return &types.QueryFeePayWalletIsEligibleResponse{
		Eligible: isEligible,
	}, nil
}

// Params returns the feepay module params
func (q Querier) Params(
	c context.Context,
	_ *types.QueryParamsRequest,
) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := q.GetParams(ctx)
	return &types.QueryParamsResponse{Params: params}, nil
}
