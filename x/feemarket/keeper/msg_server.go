package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v30/x/feemarket/types"
)

var _ types.MsgServer = (*MsgServer)(nil)

// MsgServer is the server API for x/feemarket Msg service.
type MsgServer struct {
	k *Keeper
}

// NewMsgServer returns the MsgServer implementation.
func NewMsgServer(k *Keeper) types.MsgServer {
	return &MsgServer{k}
}

// Params defines a method that updates the module's parameters. The signer of the message must
// be the module authority.
func (ms MsgServer) Params(goCtx context.Context, msg *types.MsgParams) (*types.MsgParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Authority != ms.k.GetAuthority() {
		return nil, fmt.Errorf("invalid authority to execute message")
	}

	gotParams, err := ms.k.GetParams(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting params: %w", err)
	}

	// if going from disabled -> enabled, set enabled height
	if !gotParams.Enabled && msg.Params.Enabled {
		ms.k.SetEnabledHeight(ctx, ctx.BlockHeight())
	}

	params := msg.Params
	if err := ms.k.SetParams(ctx, params); err != nil {
		return nil, fmt.Errorf("error setting params: %w", err)
	}

	newState := types.NewState(params.Window, params.MinBaseGasPrice, params.MinLearningRate)
	if err := ms.k.SetState(ctx, newState); err != nil {
		return nil, fmt.Errorf("error setting state: %w", err)
	}

	return &types.MsgParamsResponse{}, nil
}
