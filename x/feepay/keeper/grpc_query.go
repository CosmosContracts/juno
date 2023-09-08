package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v17/x/feepay/types"
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
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	contract, err := q.Keeper.GetContract(sdkCtx, req.ContractAddress)

	return &types.QueryFeePayContractResponse{
		Contract: contract,
	}, err
}

// FeePayContracts implements types.QueryServer.
func (q Querier) FeePayContracts(ctx context.Context, req *types.QueryFeePayContracts) (*types.QueryFeePayContractsResponse, error) {

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	res, err := q.Keeper.GetAllContracts(sdkCtx, req)

	if err != nil {
		return nil, err
	}

	return res, nil
}

// FeePayContractUses implements types.QueryServer.
func (q Querier) FeePayContractUses(ctx context.Context, req *types.QueryFeePayContractUses) (*types.QueryFeePayContractUsesResponse, error) {

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	uses, err := q.Keeper.GetContractUses(sdkCtx, req.ContractAddress, req.WalletAddress)

	return &types.QueryFeePayContractUsesResponse{
		Uses: uses,
	}, err
}
