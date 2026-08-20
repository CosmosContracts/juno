package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v31/x/voting-snapshot/types"
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
	if err := msg.Params.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Snapshot the previous allowlist before overwriting: addresses
	// entering or leaving the LST allowlist change both their own power
	// (zeroed vs counted) and the total, so mark the union dirty and let
	// the EndBlocker re-record everything at this height.
	var oldAllowlist []string
	if old, err := m.keeper.Params.Get(ctx); err == nil {
		oldAllowlist = old.LstAllowlist
	} else if !isNotFound(err) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := m.keeper.Params.Set(ctx, msg.Params); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	for _, list := range [][]string{oldAllowlist, msg.Params.LstAllowlist} {
		for _, bech := range list {
			addr, err := sdk.AccAddressFromBech32(bech)
			if err != nil {
				// old entries were validated when set; new ones just now
				return nil, status.Error(codes.Internal, err.Error())
			}
			if err := m.keeper.markDirty(ctx, addr); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		}
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
