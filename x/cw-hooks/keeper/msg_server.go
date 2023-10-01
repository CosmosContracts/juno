package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/CosmosContracts/juno/v17/x/cw-hooks/types"
)

var _ types.MsgServer = &msgServer{}

// msgServer is a wrapper of Keeper.
type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the x/cw-hooks MsgServer interface.
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{
		Keeper: k,
	}
}

func (k msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.authority != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

func (k msgServer) RegisterStaking(goCtx context.Context, req *types.MsgRegisterStaking) (*types.MsgRegisterStakingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: (GetWasmKeeper should be a pointer)
	// contract, err := sdk.AccAddressFromBech32(req.Contract.ContractAddress)
	// if err != nil {
	// 	return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid contract address (%s)", err)
	// }

	// sender, err := sdk.AccAddressFromBech32(req.Contract.RegisterAddress)
	// if err != nil {
	// 	return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address (%s)", err)
	// }

	// if k.IsStakingContractRegistered(ctx, contract) {
	// 	return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "staking contract already registered")
	// }

	// contractInfo := k.GetWasmKeeper().GetContractInfo(ctx, contract)

	// // TODO: validate this / move to its own function
	// if contractInfo.Creator != "" && contractInfo.Creator != sender.String() {
	// 	return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "sender is not the contract creator")
	// } else if contractInfo.Admin != "" && contractInfo.Admin != sender.String() {
	// 	return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "sender is not the contract admin")
	// }

	k.SetStakingContract(ctx, req.Contract)

	return &types.MsgRegisterStakingResponse{}, nil
}
