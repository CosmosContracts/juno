package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/CosmosContracts/juno/v30/x/voting-snapshot/types"
)

// MsgServer adapts Keeper into the proto-generated types.MsgServer
// interface. Currently a single message: governance-driven UpdateParams.
type MsgServer struct {
	keeper Keeper
}

func NewMsgServer(k Keeper) types.MsgServer {
	return &MsgServer{keeper: k}
}

var _ types.MsgServer = &MsgServer{}

func (m *MsgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty message")
	}
	if msg.Authority != m.keeper.Authority() {
		return nil, status.Errorf(codes.PermissionDenied, "expected authority %s, got %s", m.keeper.Authority(), msg.Authority)
	}
	if err := m.keeper.Params.Set(ctx, msg.Params); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.MsgUpdateParamsResponse{}, nil
}
