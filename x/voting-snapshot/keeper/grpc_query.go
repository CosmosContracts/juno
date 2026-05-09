package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v30/x/voting-snapshot/types"
)

// QueryServer adapts Keeper into the proto-generated types.QueryServer
// interface. The wasmbinding surface mirrors these methods directly;
// gRPC consumers (CLI, indexers, REST gateway) use this path.
type QueryServer struct {
	keeper Keeper
}

func NewQueryServer(k Keeper) types.QueryServer {
	return &QueryServer{keeper: k}
}

var _ types.QueryServer = &QueryServer{}

func (q *QueryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	params, err := q.keeper.Params.Get(ctx)
	if err != nil {
		if isNotFound(err) {
			return &types.QueryParamsResponse{Params: types.DefaultParams()}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryParamsResponse{Params: params}, nil
}

func (q *QueryServer) VotingPowerAt(ctx context.Context, req *types.QueryVotingPowerAtRequest) (*types.QueryVotingPowerAtResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address: %s", err)
	}
	power, err := q.keeper.VotingPowerAt(ctx, addr, req.Height)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryVotingPowerAtResponse{Power: power.String()}, nil
}

func (q *QueryServer) TotalVotingPowerAt(ctx context.Context, req *types.QueryTotalVotingPowerAtRequest) (*types.QueryTotalVotingPowerAtResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	power, err := q.keeper.TotalVotingPowerAt(ctx, req.Height)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryTotalVotingPowerAtResponse{Power: power.String()}, nil
}

func (q *QueryServer) VotingPowerOverRange(ctx context.Context, req *types.QueryVotingPowerOverRangeRequest) (*types.QueryVotingPowerOverRangeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address: %s", err)
	}
	rows, err := q.keeper.VotingPowerOverRange(ctx, addr, req.FromHeight, req.ToHeight)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := make([]types.HeightPower, 0, len(rows))
	for _, r := range rows {
		out = append(out, types.HeightPower{Height: r.Height, Power: r.Power.String()})
	}
	return &types.QueryVotingPowerOverRangeResponse{Rows: out}, nil
}
