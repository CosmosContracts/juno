package keeper

import (
	"context"
	"slices"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CosmosContracts/juno/v30/x/drip/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the x/mint MsgServer interface.
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{
		Keeper: k,
	}
}

// DistributeTokens distribute tokens to all stakers at the next block
func (ms msgServer) DistributeTokens(
	goCtx context.Context,
	msg *types.MsgDistributeTokens,
) (*types.MsgDistributeTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.SenderAddress == "" {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.SenderAddress); err != nil {
		return nil, errorsmod.Wrapf(err, "invalid sender address: %s", err.Error())
	}

	if msg.Amount == nil || msg.Amount.Empty() {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "invalid coins: %s", msg.Amount.String())
	}

	if !msg.Amount.IsValid() {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "invalid coins: %s", msg.Amount.String())
	}

	params := ms.GetParams(ctx)
	if !params.EnableDrip {
		return nil, types.ErrDripDisabled
	}

	// Check if sender is allowed
	authorized := slices.Contains(params.AllowedAddresses, msg.SenderAddress)

	if !authorized {
		return nil, types.ErrDripNotAllowed
	}

	// Get sender
	sender, err := sdk.AccAddressFromBech32(msg.SenderAddress)
	if err != nil {
		return nil, err
	}

	if err := ms.SendCoinsFromAccountToFeeCollector(ctx, sender, msg.Amount); err != nil {
		return nil, err
	}

	return &types.MsgDistributeTokensResponse{}, nil
}

func (ms msgServer) UpdateParams(ctx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if ms.authority != req.Authority {
		return nil, errorsmod.Wrapf(sdkerrors.ErrorInvalidSigner, "expected %s, got %s", ms.authority, req.Authority)
	}
	err := req.Params.Validate()
	if err != nil {
		return nil, err
	}

	if err := ms.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
