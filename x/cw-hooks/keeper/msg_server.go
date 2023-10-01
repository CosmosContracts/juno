package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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

	if err := k.handleContractRegister(ctx, req.ContractAddress, types.KeyPrefixStaking, "staking"); err != nil {
		return nil, err
	}

	return &types.MsgRegisterStakingResponse{}, nil
}

func (k msgServer) RegisterGovernance(goCtx context.Context, req *types.MsgRegisterGovernance) (*types.MsgRegisterGovernanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.handleContractRegister(ctx, req.ContractAddress, types.KeyPrefixGov, "governance"); err != nil {
		return nil, err
	}

	return &types.MsgRegisterGovernanceResponse{}, nil
}

func (k msgServer) handleContractRegister(ctx sdk.Context, contractAddr string, keyPrefix []byte, prefixModuleName string) error {
	contract, err := sdk.AccAddressFromBech32(contractAddr)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid contract address (%s)", err)
	}

	// if k.GetWasmKeeper().HasContractInfo(ctx, contract) {
	// 	return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "this contract is not found in the wasm module store")
	// }

	if k.IsContractRegistered(ctx, keyPrefix, contract) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "contract already registered for %s", prefixModuleName)
	}

	// contractInfo := k.GetWasmKeeper().GetContractInfo(ctx, contract)
	// if contractInfo.Creator != "" && contractInfo.Creator != sender.String() {
	// 	return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "sender is not the contract creator")
	// } else if contractInfo.Admin != "" && contractInfo.Admin != sender.String() {
	// 	return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "sender is not the contract admin")
	// }

	k.SetContract(ctx, keyPrefix, contract)

	return nil
}
