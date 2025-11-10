package keeper

import (
	"context"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CosmosContracts/juno/v30/x/cw-hooks/types"
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

func (k msgServer) UpdateParams(ctx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.authority != req.Authority {
		return nil, errorsmod.Wrapf(sdkerrors.ErrorInvalidSigner, "expected %s, got %s", k.authority, req.Authority)
	}

	if _, err := sdk.AccAddressFromBech32(req.Authority); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid authority address")
	}

	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

func (k msgServer) RegisterContract(ctx context.Context, req *types.MsgRegisterContract) (*types.MsgRegisterContractResponse, error) {
	modulePrefix, err := types.ModulePrefixFromModule(req.Module)
	if err != nil {
		return nil, err
	}

	if _, err := sdk.AccAddressFromBech32(req.ContractAddress); err != nil {
		return nil, errorsmod.Wrap(err, "invalid contract address")
	}

	if err := k.handleContractRegister(ctx, req.SenderAddress, req.ContractAddress, modulePrefix); err != nil {
		return nil, err
	}

	return &types.MsgRegisterContractResponse{}, nil
}

func (k msgServer) UnregisterContract(ctx context.Context, req *types.MsgUnregisterContract) (*types.MsgUnregisterContractResponse, error) {
	modulePrefix, err := types.ModulePrefixFromModule(req.Module)
	if err != nil {
		return nil, err
	}

	if _, err := sdk.AccAddressFromBech32(req.SenderAddress); err != nil {
		return nil, errorsmod.Wrap(err, "invalid sender address")
	}

	if _, err := sdk.AccAddressFromBech32(req.ContractAddress); err != nil {
		return nil, errorsmod.Wrap(err, "invalid contract address")
	}

	if err := k.handleContractRemoval(ctx, req.SenderAddress, req.ContractAddress, modulePrefix); err != nil {
		return nil, err
	}

	return &types.MsgUnregisterContractResponse{}, nil
}

func (k msgServer) isContractSenderAuthorized(ctx context.Context, sender string, contract sdk.AccAddress) error {
	if ok := k.GetWasmKeeper().HasContractInfo(ctx, contract); !ok {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "contract does not exist: %s", contract)
	}

	contractInfo := k.GetWasmKeeper().GetContractInfo(ctx, contract)

	if contractInfo.Creator != "" && contractInfo.Creator != sender {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "sender is not the contract creator")
	} else if contractInfo.Admin != "" && contractInfo.Admin != sender {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "sender is not the contract admin")
	}

	return nil
}

func (k msgServer) handleContractRegister(ctx context.Context, sender string, contractAddr string, key collections.Prefix) error {
	addr, err := sdk.AccAddressFromBech32(contractAddr)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid contract address (%s)", err)
	}

	if ok, err := k.IsContractRegistered(ctx, key, addr); err != nil {
		return err
	} else if ok {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "contract already registered for %s", key)
	}

	if err := k.isContractSenderAuthorized(ctx, sender, addr); err != nil {
		return err
	}

	contract := types.ContractInfo{
		ContractAddress: contractAddr,
		FailureCounter:  0,
	}

	return k.SetContract(ctx, key, contract)
}

func (k msgServer) handleContractRemoval(ctx context.Context, sender, contractAddr string, key collections.Prefix) error {
	addr, err := sdk.AccAddressFromBech32(contractAddr)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid contract address (%s)", err)
	}

	if ok, err := k.IsContractRegistered(ctx, key, addr); err != nil {
		return err
	} else if !ok {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "contract is not registered for %s", key)
	}

	if err := k.isContractSenderAuthorized(ctx, sender, addr); err != nil {
		return err
	}

	return k.DeleteContract(ctx, key, addr)
}
