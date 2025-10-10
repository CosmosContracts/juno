package keeper

import (
	"context"
	"errors"

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

// UpdateParams defines a method that updates the module's parameters. The signer of the message must
// be the module authority.
func (ms MsgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Authority != ms.k.GetAuthority() {
		return nil, errors.New("invalid authority to execute message")
	}

	gotParams, err := ms.k.GetParams(ctx)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	// if going from disabled -> enabled, set enabled height
	if !gotParams.Enabled && msg.Params.Enabled {
		ms.k.SetEnabledHeight(ctx, ctx.BlockHeight())
	}

	params := msg.Params
	if err := ms.k.SetParams(ctx, params); err != nil {
		return nil, errors.New(err.Error())
	}

	newState := types.NewState(params.Window, params.MinBaseGasPrice, params.MinLearningRate)
	if err := ms.k.SetState(ctx, newState); err != nil {
		return nil, errors.New(err.Error())
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
