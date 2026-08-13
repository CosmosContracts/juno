package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CosmosContracts/juno/v31/x/feemarket/types"
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
		return nil, errorsmod.Wrapf(sdkerrors.ErrorInvalidSigner, "expected %s, got %s", ms.k.GetAuthority(), msg.Authority)
	}

	// Re-validate here as a defense in depth: params that fail validation
	// (zero window, empty fee denom, nil decimals, inverted learning-rate
	// bounds, ...) would panic or deterministically error in the ante/post
	// handlers and EndBlock — halting the chain.
	params := msg.Params
	if err := params.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "invalid params")
	}

	gotParams, err := ms.k.GetParams(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to get params")
	}
	if gotParams.FeeDenom != params.FeeDenom && ms.k.feePayLiabilities != nil && ms.k.feePayLiabilities.HasOutstandingBalances(ctx) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "cannot change fee denom while FeePay has outstanding FeePay balances")
	}

	// if going from disabled -> enabled, set enabled height
	if !gotParams.Enabled && msg.Params.Enabled {
		ms.k.SetEnabledHeight(ctx, ctx.BlockHeight())
	}

	if err := ms.k.SetParams(ctx, params); err != nil {
		return nil, errorsmod.Wrap(err, "failed to set params")
	}

	newState := types.NewState(params.Window, params.MinBaseGasPrice, params.MinLearningRate)
	if err := newState.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "invalid state derived from params")
	}
	if err := ms.k.SetState(ctx, newState); err != nil {
		return nil, errorsmod.Wrap(err, "failed to set state")
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
