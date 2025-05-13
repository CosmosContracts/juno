package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	globalerrors "github.com/CosmosContracts/juno/v29/app/helpers"
	"github.com/CosmosContracts/juno/v29/x/feepay/types"
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

// Register a new fee pay contract.
func (ms msgServer) RegisterFeePayContract(ctx context.Context, msg *types.MsgRegisterFeePayContract) (*types.MsgRegisterFeePayContractResponse, error) {
	// Prevent client from overriding initial contract balance of zero
	msg.FeePayContract.Balance = uint64(0)
	return &types.MsgRegisterFeePayContractResponse{}, ms.RegisterContract(ctx, msg)
}

func (ms msgServer) UnregisterFeePayContract(ctx context.Context, msg *types.MsgUnregisterFeePayContract) (*types.MsgUnregisterFeePayContractResponse, error) {
	return &types.MsgUnregisterFeePayContractResponse{}, ms.UnregisterContract(ctx, msg)
}

// FundFeePayContract funds a contract with the given amount of tokens.
func (ms msgServer) FundFeePayContract(ctx context.Context, msg *types.MsgFundFeePayContract) (*types.MsgFundFeePayContractResponse, error) {
	if _, err := sdk.AccAddressFromBech32(msg.ContractAddress); err != nil {
		return nil, err
	}

	if len(msg.Amount) != 1 {
		return nil, types.ErrInvalidJunoFundAmount
	}

	// Get the contract
	contract, err := ms.GetContract(ctx, msg.ContractAddress)
	if err != nil {
		return nil, err
	}

	// Validate sender address
	senderAddr, err := sdk.AccAddressFromBech32(msg.SenderAddress)
	if err != nil {
		return nil, errorsmod.Wrapf(globalerrors.ErrInvalidAddress, "invalid sender address: %s", msg.SenderAddress)
	}

	return &types.MsgFundFeePayContractResponse{}, ms.FundContract(ctx, contract, senderAddr, msg.Amount)
}

// Update the wallet limit of a fee pay contract.
func (ms msgServer) UpdateFeePayContractWalletLimit(ctx context.Context, msg *types.MsgUpdateFeePayContractWalletLimit) (*types.MsgUpdateFeePayContractWalletLimitResponse, error) {
	if _, err := sdk.AccAddressFromBech32(msg.SenderAddress); err != nil {
		return nil, err
	}

	if _, err := sdk.AccAddressFromBech32(msg.ContractAddress); err != nil {
		return nil, err
	}

	// Get the contract
	contract, err := ms.GetContract(ctx, msg.ContractAddress)
	if err != nil {
		return nil, err
	}

	if msg.WalletLimit > 1_000_000 {
		return nil, errorsmod.Wrapf(types.ErrInvalidWalletLimit, "invalid wallet limit: %d", msg.WalletLimit)
	}

	return &types.MsgUpdateFeePayContractWalletLimitResponse{}, ms.UpdateContractWalletLimit(ctx, contract, msg.SenderAddress, msg.WalletLimit)
}

// UpdateParams updates the parameters of the module.
func (ms msgServer) UpdateParams(ctx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if ms.authority != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}

	if err := ms.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
