package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

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
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid authority; expected %s, got %s", ms.k.GetAuthority(), msg.Authority)
	}

	gotParams, err := ms.k.GetParams(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to get params")
	}

	// if going from disabled -> enabled, set enabled height
	if !gotParams.Enabled && msg.Params.Enabled {
		ms.k.SetEnabledHeight(ctx, ctx.BlockHeight())
	}

	params := msg.Params
	if err := ms.k.SetParams(ctx, params); err != nil {
		return nil, errorsmod.Wrap(err, "failed to set params")
	}

	newState := types.NewState(params.Window, params.MinBaseGasPrice, params.MinLearningRate)
	if err := ms.k.SetState(ctx, newState); err != nil {
		return nil, errorsmod.Wrap(err, "failed to set state")
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
