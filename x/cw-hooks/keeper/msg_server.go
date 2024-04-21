package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/CosmosContracts/juno/v22/x/cw-hooks/types"
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

	if err := k.handleContractRegister(ctx, req.RegisterAddress, req.ContractAddress, types.KeyPrefixStaking, "staking"); err != nil {
		return nil, err
	}

	return &types.MsgRegisterStakingResponse{}, nil
}

func (k msgServer) RegisterGovernance(goCtx context.Context, req *types.MsgRegisterGovernance) (*types.MsgRegisterGovernanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.handleContractRegister(ctx, req.RegisterAddress, req.ContractAddress, types.KeyPrefixGov, "governance"); err != nil {
		return nil, err
	}

	return &types.MsgRegisterGovernanceResponse{}, nil
}

func (k msgServer) UnregisterGovernance(goCtx context.Context, req *types.MsgUnregisterGovernance) (*types.MsgUnregisterGovernanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.handleContractRemoval(ctx, req.RegisterAddress, req.ContractAddress, types.KeyPrefixGov, "governance"); err != nil {
		return nil, err
	}

	return &types.MsgUnregisterGovernanceResponse{}, nil
}

func (k msgServer) UnregisterStaking(goCtx context.Context, req *types.MsgUnregisterStaking) (*types.MsgUnregisterStakingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.handleContractRemoval(ctx, req.RegisterAddress, req.ContractAddress, types.KeyPrefixStaking, "staking"); err != nil {
		return nil, err
	}

	return &types.MsgUnregisterStakingResponse{}, nil
}

func (k msgServer) isContractSenderAuthorized(ctx sdk.Context, sender string, contract sdk.AccAddress) error {
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

func (k msgServer) handleContractRegister(ctx sdk.Context, sender, contractAddr string, keyPrefix []byte, prefixModuleName string) error {
	contract, err := sdk.AccAddressFromBech32(contractAddr)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid contract address (%s)", err)
	}

	if k.IsContractRegistered(ctx, keyPrefix, contract) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "contract already registered for %s", prefixModuleName)
	}

	if err := k.isContractSenderAuthorized(ctx, sender, contract); err != nil {
		return err
	}

	k.SetContract(ctx, keyPrefix, contract)

	return nil
}

func (k msgServer) handleContractRemoval(ctx sdk.Context, sender, contractAddr string, keyPrefix []byte, prefixModuleName string) error {
	contract, err := sdk.AccAddressFromBech32(contractAddr)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid contract address (%s)", err)
	}

	if !k.IsContractRegistered(ctx, keyPrefix, contract) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "contract is not registered for %s", prefixModuleName)
	}

	if err := k.isContractSenderAuthorized(ctx, sender, contract); err != nil {
		return err
	}

	k.DeleteContract(ctx, keyPrefix, contract)

	return nil
}
