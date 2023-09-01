package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/CosmosContracts/juno/v17/x/feepay/types"
)

var _ types.MsgServer = &Keeper{}

func (k Keeper) RegisterFeePayContract(goCtx context.Context, msg *types.MsgRegisterFeePayContract) (*types.MsgRegisterFeePayContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.RegisterContract(ctx, *msg.Contract)
	return &types.MsgRegisterFeePayContractResponse{}, nil
}

// FundFeePayContract funds a contract with the given amount of tokens.
func (k Keeper) FundFeePayContract(goCtx context.Context, msg *types.MsgFundFeePayContract) (*types.MsgFundFeePayContractResponse, error) {

	// TODO: IMPLEMENT

	return nil, nil
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
