package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	globalerrors "github.com/CosmosContracts/juno/v25/app/helpers"
	"github.com/CosmosContracts/juno/v25/x/feepay/types"
)

var _ types.MsgServer = &Keeper{}

// Register a new fee pay contract.
func (k Keeper) RegisterFeePayContract(goCtx context.Context, msg *types.MsgRegisterFeePayContract) (*types.MsgRegisterFeePayContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Prevent client from overriding initial contract balance of zero
	msg.FeePayContract.Balance = uint64(0)
	return &types.MsgRegisterFeePayContractResponse{}, k.RegisterContract(ctx, msg)
}

func (k Keeper) UnregisterFeePayContract(goCtx context.Context, msg *types.MsgUnregisterFeePayContract) (*types.MsgUnregisterFeePayContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.MsgUnregisterFeePayContractResponse{}, k.UnregisterContract(ctx, msg)
}

// FundFeePayContract funds a contract with the given amount of tokens.
func (k Keeper) FundFeePayContract(goCtx context.Context, msg *types.MsgFundFeePayContract) (*types.MsgFundFeePayContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get the contract
	contract, err := k.GetContract(ctx, msg.ContractAddress)
	if err != nil {
		return nil, err
	}

	// Validate sender address
	senderAddr, err := sdk.AccAddressFromBech32(msg.SenderAddress)
	if err != nil {
		return nil, errorsmod.Wrapf(globalerrors.ErrInvalidAddress, "invalid sender address: %s", msg.SenderAddress)
	}

	return &types.MsgFundFeePayContractResponse{}, k.FundContract(ctx, contract, senderAddr, msg.Amount)
}

// Update the wallet limit of a fee pay contract.
func (k Keeper) UpdateFeePayContractWalletLimit(goCtx context.Context, msg *types.MsgUpdateFeePayContractWalletLimit) (*types.MsgUpdateFeePayContractWalletLimitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get the contract
	contract, err := k.GetContract(ctx, msg.ContractAddress)
	if err != nil {
		return nil, err
	}

	if msg.WalletLimit > 1_000_000 {
		return nil, errorsmod.Wrapf(types.ErrInvalidWalletLimit, "invalid wallet limit: %d", msg.WalletLimit)
	}

	return &types.MsgUpdateFeePayContractWalletLimitResponse{}, k.UpdateContractWalletLimit(ctx, contract, msg.SenderAddress, msg.WalletLimit)
}

// UpdateParams updates the parameters of the module.
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
