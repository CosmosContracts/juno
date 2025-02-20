package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/store/prefix"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CosmosContracts/juno/v28/x/feeshare/types"
)

var _ types.QueryServer = queryServer{}

func NewQueryServerImpl(k Keeper) types.QueryServer {
	return queryServer{k}
}

type queryServer struct {
	k Keeper
}

// FeeShares returns all FeeShares that have been registered for fee distribution
func (q queryServer) FeeShares(
	ctx context.Context,
	req *types.QueryFeeSharesRequest,
) (*types.QueryFeeSharesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	var feeshares []types.FeeShare
	key := runtime.KVStoreAdapter(q.k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(key, types.KeyPrefixFeeShare)

	pageRes, err := query.Paginate(store, req.Pagination, func(_, value []byte) error {
		var feeshare types.FeeShare
		if err := q.k.cdc.Unmarshal(value, &feeshare); err != nil {
			return err
		}
		feeshares = append(feeshares, feeshare)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryFeeSharesResponse{
		Feeshare:   feeshares,
		Pagination: pageRes,
	}, nil
}

// FeeShare returns the FeeShare that has been registered for fee distribution for a given
// contract
func (q queryServer) FeeShare(
	ctx context.Context,
	req *types.QueryFeeShareRequest,
) (*types.QueryFeeShareResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// check if the contract is a non-zero hex address
	contract, err := sdk.AccAddressFromBech32(req.ContractAddress)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"invalid format for contract %s, should be bech32 ('juno...')", req.ContractAddress,
		)
	}

	feeshare, found := q.k.GetFeeShare(ctx, contract)
	if !found {
		return nil, status.Errorf(
			codes.NotFound,
			"fees registered contract '%s'",
			req.ContractAddress,
		)
	}

	return &types.QueryFeeShareResponse{Feeshare: feeshare}, nil
}

// Params returns the fees module params
func (q queryServer) Params(
	ctx context.Context,
	_ *types.QueryParamsRequest,
) (*types.QueryParamsResponse, error) {
	params := q.k.GetParams(ctx)
	return &types.QueryParamsResponse{Params: params}, nil
}

// DeployerFeeShares returns all contracts that have been registered for fee
// distribution by a given deployer
func (q queryServer) DeployerFeeShares( // nolint: dupl
	ctx context.Context,
	req *types.QueryDeployerFeeSharesRequest,
) (*types.QueryDeployerFeeSharesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	deployer, err := sdk.AccAddressFromBech32(req.DeployerAddress)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"invalid format for deployer %s, should be bech32 ('juno...')", req.DeployerAddress,
		)
	}

	var contracts []string
	key := runtime.KVStoreAdapter(q.k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(
		key,
		types.GetKeyPrefixDeployer(deployer),
	)

	pageRes, err := query.Paginate(store, req.Pagination, func(key, _ []byte) error {
		contracts = append(contracts, sdk.AccAddress(key).String())
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryDeployerFeeSharesResponse{
		ContractAddresses: contracts,
		Pagination:        pageRes,
	}, nil
}

// WithdrawerFeeShares returns all fees for a given withdraw address
func (q queryServer) WithdrawerFeeShares( // nolint: dupl
	ctx context.Context,
	req *types.QueryWithdrawerFeeSharesRequest,
) (*types.QueryWithdrawerFeeSharesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	deployer, err := sdk.AccAddressFromBech32(req.WithdrawerAddress)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"invalid format for withdraw addr %s, should be bech32 ('juno...')", req.WithdrawerAddress,
		)
	}

	var contracts []string
	key := runtime.KVStoreAdapter(q.k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(
		key,
		types.GetKeyPrefixWithdrawer(deployer),
	)

	pageRes, err := query.Paginate(store, req.Pagination, func(key, _ []byte) error {
		contracts = append(contracts, sdk.AccAddress(key).String())

		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryWithdrawerFeeSharesResponse{
		ContractAddresses: contracts,
		Pagination:        pageRes,
	}, nil
}
