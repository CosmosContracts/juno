package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CosmosContracts/juno/v12/x/feeshare/types"
)

var _ types.MsgServer = &Keeper{}

// CheckIfDeployerIsContractAdmin ensures the deployer is the contract's admin for all msg_server revenue functions.
// If not, it returns an error & we know the user is not authorized to perform the action.
func (k Keeper) CheckIfDeployerIsContractAdmin(ctx sdk.Context, contract sdk.AccAddress, deployer string) (deployerAddr sdk.AccAddress, errr error) {
	deployerAddr, err := sdk.AccAddressFromBech32(deployer)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid deployer address %s", deployer)
	}

	info := k.wasmKeeper.GetContractInfo(ctx, contract)

	if len(info.Admin) == 0 {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "contract %s has no admin set", contract)
	}

	if info.Admin != deployer {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "you are not an admin of this contract %s", deployer)
	}

	return deployerAddr, nil
}

// RegisterRevenue registers a contract to receive transaction fees
func (k Keeper) RegisterRevenue(
	goCtx context.Context,
	msg *types.MsgRegisterRevenue,
) (*types.MsgRegisterRevenueResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params := k.GetParams(ctx)
	if !params.EnableRevenue {
		return nil, types.ErrRevenueDisabled
	}

	// Get Contract
	contract, err := sdk.AccAddressFromBech32(msg.ContractAddress)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid contract address (%s)", err)
	}

	// Check if contract is already registered
	if k.IsRevenueRegistered(ctx, contract) {
		return nil, sdkerrors.Wrapf(types.ErrRevenueAlreadyRegistered, "contract is already registered %s", contract)
	}

	// Check that the person who signed the message is the wasm contract admin, if so return the deployer address
	deployer, err := k.CheckIfDeployerIsContractAdmin(ctx, contract, msg.DeployerAddress)
	if err != nil {
		return nil, err
	}

	// Get the withdraw address of the contract
	withdrawer, err := sdk.AccAddressFromBech32(msg.WithdrawerAddress)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid withdrawer address %s", msg.WithdrawerAddress)
	}

	// prevent storing the same address for deployer and withdrawer
	revenue := types.NewRevenue(contract, deployer, withdrawer)
	k.SetRevenue(ctx, revenue)
	k.SetDeployerMap(ctx, deployer, contract)

	if len(withdrawer.String()) != 0 {
		k.SetWithdrawerMap(ctx, withdrawer, contract)
	}

	k.Logger(ctx).Debug(
		"registering contract for transaction fees",
		"contract", msg.ContractAddress,
		"deployer", msg.DeployerAddress,
		"withdraw", msg.WithdrawerAddress,
	)

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeRegisterRevenue,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.DeployerAddress),
				sdk.NewAttribute(types.AttributeKeyContract, msg.ContractAddress),
				sdk.NewAttribute(types.AttributeKeyWithdrawerAddress, msg.WithdrawerAddress),
			),
		},
	)

	return &types.MsgRegisterRevenueResponse{}, nil
}

// UpdateRevenue updates the withdraw address of a given Revenue. If the given
// withdraw address is empty or the same as the deployer address, the withdraw
// address is removed.
func (k Keeper) UpdateRevenue(
	goCtx context.Context,
	msg *types.MsgUpdateRevenue,
) (*types.MsgUpdateRevenueResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params := k.GetParams(ctx)
	if !params.EnableRevenue {
		return nil, types.ErrRevenueDisabled
	}

	contract, err := sdk.AccAddressFromBech32(msg.ContractAddress)
	if err != nil {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidAddress,
			"invalid contract address (%s)", err,
		)
	}

	revenue, found := k.GetRevenue(ctx, contract)
	if !found {
		return nil, sdkerrors.Wrapf(
			types.ErrRevenueContractNotRegistered,
			"contract %s is not registered", msg.ContractAddress,
		)
	}

	// revenue with the given withdraw address is already registered
	if msg.WithdrawerAddress == revenue.WithdrawerAddress {
		return nil, sdkerrors.Wrapf(types.ErrRevenueAlreadyRegistered, "revenue with withdraw address %s is already registered", msg.WithdrawerAddress)
	}

	// Check that the person who signed the message is the wasm contract admin, if so return the deployer address
	_, err = k.CheckIfDeployerIsContractAdmin(ctx, contract, msg.DeployerAddress)
	if err != nil {
		return nil, err
	}

	withdrawAddr, err := sdk.AccAddressFromBech32(revenue.WithdrawerAddress)
	if err != nil {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidAddress,
			"invalid withdrawer address (%s)", err,
		)
	}

	// only delete withdrawer map if is not default
	if revenue.WithdrawerAddress != "" {
		k.DeleteWithdrawerMap(ctx, withdrawAddr, contract)
	}

	// only add withdrawer map if new entry is not default
	if msg.WithdrawerAddress != "" {
		k.SetWithdrawerMap(
			ctx,
			withdrawAddr,
			contract,
		)
	}
	// update revenue
	revenue.WithdrawerAddress = msg.WithdrawerAddress
	k.SetRevenue(ctx, revenue)

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeUpdateRevenue,
				sdk.NewAttribute(types.AttributeKeyContract, msg.ContractAddress),
				sdk.NewAttribute(sdk.AttributeKeySender, msg.DeployerAddress),
				sdk.NewAttribute(types.AttributeKeyWithdrawerAddress, msg.WithdrawerAddress),
			),
		},
	)

	return &types.MsgUpdateRevenueResponse{}, nil
}

// CancelRevenue deletes the Revenue for a given contract
func (k Keeper) CancelRevenue(
	goCtx context.Context,
	msg *types.MsgCancelRevenue,
) (*types.MsgCancelRevenueResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params := k.GetParams(ctx)
	if !params.EnableRevenue {
		return nil, types.ErrRevenueDisabled
	}

	contract, err := sdk.AccAddressFromBech32(msg.ContractAddress)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid contract address (%s)", err)
	}

	fee, found := k.GetRevenue(ctx, contract)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrRevenueContractNotRegistered, "contract %s is not registered", msg.ContractAddress)
	}

	// Check that the person who signed the message is the wasm contract admin, if so return the deployer address
	_, err = k.CheckIfDeployerIsContractAdmin(ctx, contract, msg.DeployerAddress)
	if err != nil {
		return nil, err
	}

	k.DeleteRevenue(ctx, fee)
	k.DeleteDeployerMap(
		ctx,
		fee.GetDeployerAddr(),
		contract,
	)

	// delete entry from withdrawer map if not default
	if fee.WithdrawerAddress != "" {
		k.DeleteWithdrawerMap(
			ctx,
			fee.GetWithdrawerAddr(),
			contract,
		)
	}

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeCancelRevenue,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.DeployerAddress),
				sdk.NewAttribute(types.AttributeKeyContract, msg.ContractAddress),
			),
		},
	)

	return &types.MsgCancelRevenueResponse{}, nil
}
