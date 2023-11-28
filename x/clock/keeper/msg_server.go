package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/CosmosContracts/juno/v18/x/clock/types"
)

var _ types.MsgServer = &msgServer{}

// msgServer is a wrapper of Keeper.
type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the x/clock MsgServer interface.
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{
		Keeper: k,
	}
}

// RegisterClockContract handles incoming transactions to register clock contracts.
func (k msgServer) RegisterClockContract(goCtx context.Context, req *types.MsgRegisterClockContract) (*types.MsgRegisterClockContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate sender address
	if _, err := sdk.AccAddressFromBech32(req.SenderAddress); err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid sender address: %s", req.SenderAddress)
	}

	// Validate contract address
	if _, err := sdk.AccAddressFromBech32(req.ContractAddress); err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid contract address: %s", req.ContractAddress)
	}

	return &types.MsgRegisterClockContractResponse{}, k.RegisterContract(ctx, req.SenderAddress, req.ContractAddress)
}

// UnregisterClockContract handles incoming transactions to unregister clock contracts.
func (k msgServer) UnregisterClockContract(goCtx context.Context, req *types.MsgUnregisterClockContract) (*types.MsgUnregisterClockContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate sender address
	if _, err := sdk.AccAddressFromBech32(req.SenderAddress); err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid sender address: %s", req.SenderAddress)
	}

	// Validate contract address
	if _, err := sdk.AccAddressFromBech32(req.ContractAddress); err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid contract address: %s", req.ContractAddress)
	}

	return &types.MsgUnregisterClockContractResponse{}, k.UnregisterContract(ctx, req.SenderAddress, req.ContractAddress)
}

// UnjailClockContract handles incoming transactions to unjail clock contracts.
func (k msgServer) UnjailClockContract(goCtx context.Context, req *types.MsgUnjailClockContract) (*types.MsgUnjailClockContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate sender address
	if _, err := sdk.AccAddressFromBech32(req.SenderAddress); err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid sender address: %s", req.SenderAddress)
	}

	// Validate contract address
	if _, err := sdk.AccAddressFromBech32(req.ContractAddress); err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid contract address: %s", req.ContractAddress)
	}

	return &types.MsgUnjailClockContractResponse{}, k.SetJailStatusBySender(ctx, req.SenderAddress, req.ContractAddress, false)
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
