package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/CosmosContracts/juno/v24/x/drip/types"
)

var _ types.MsgServer = &Keeper{}

// DistributeTokens distribute tokens to all stakers at the next block
func (k Keeper) DistributeTokens(
	goCtx context.Context,
	msg *types.MsgDistributeTokens,
) (*types.MsgDistributeTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	params := k.GetParams(ctx)
	if !params.EnableDrip {
		return nil, types.ErrDripDisabled
	}

	// Check if sender is allowed
	authorized := false
	for _, addr := range params.AllowedAddresses {
		if msg.SenderAddress == addr {
			authorized = true
			break
		}
	}

	if !authorized {
		return nil, types.ErrDripNotAllowed
	}

	// Get sender
	sender, err := sdk.AccAddressFromBech32(msg.SenderAddress)
	if err != nil {
		return nil, err
	}

	if err := k.SendCoinsFromAccountToFeeCollector(ctx, sender, msg.Amount); err != nil {
		return nil, err
	}

	return &types.MsgDistributeTokensResponse{}, nil
}

func (k Keeper) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.authority != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
