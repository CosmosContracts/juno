package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CosmosContracts/juno/v12/x/feeshare/types"
)

// type msgServer struct {
// 	keeper *Keeper
// }

var _ types.MsgServer = &Keeper{}

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

	contract, err := sdk.AccAddressFromBech32(msg.ContractAddress)
	if err != nil {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidAddress,
			"invalid contract address (%s)", err,
		)
	}

	if k.IsRevenueRegistered(ctx, contract) {
		return nil, sdkerrors.Wrapf(
			types.ErrRevenueAlreadyRegistered,
			"contract is already registered %s", contract,
		)
	}

	deployer, err := sdk.AccAddressFromBech32(msg.DeployerAddress)
	if err != nil {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidAddress,
			"invalid deployer address %s", msg.DeployerAddress,
		)
	}

	// we require deployer to be the --from key, so this would never hit right?
	// deployerAccount := k.accountKeeper.GetAccount(ctx, deployer)
	// if deployerAccount == nil || deployerAccount.GetSequence() == 0 {
	// 	return nil, sdkerrors.Wrapf(
	// 		sdkerrors.ErrNotFound,
	// 		"deployer account not found %s", msg.DeployerAddress,
	// 	)
	// }

	// TODO: Check that the admin of the contract is the one requesting the funds. If not, unauthorized.
	info := k.wasmKeeper.GetContractInfo(ctx, contract)
	// Do we want to do anything with the Creator if no admin is set?
	if len(info.Admin) == 0 {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrUnauthorized,
			"contract %s has no admin set", contract,
		)
	}

	if info.Admin != deployer.String() {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrUnauthorized,
			"you are not an admin of this contract %s", deployer,
		)
	}

	withdrawer, err := sdk.AccAddressFromBech32(msg.WithdrawerAddress)
	if err != nil {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidAddress,
			"invalid withdrawer address %s", msg.WithdrawerAddress,
		)
	}

	// prevent storing the same address for deployer and withdrawer
	revenue := types.NewRevenue(contract, deployer, withdrawer)
	k.SetRevenue(ctx, revenue)
	k.SetDeployerMap(ctx, deployer, contract)

	// The effective withdrawer is the withdraw address that is stored after the
	// revenue registration is completed. It defaults to the deployer address if
	// the withdraw address in the msg is omitted. When omitted, the withdraw map
	// doesn't need to be set.
	effectiveWithdrawer := msg.DeployerAddress

	if len(withdrawer.String()) != 0 {
		k.SetWithdrawerMap(ctx, withdrawer, contract)
		effectiveWithdrawer = msg.WithdrawerAddress
	}

	k.Logger(ctx).Debug(
		"registering contract for transaction fees",
		"contract", msg.ContractAddress, "deployer", msg.DeployerAddress,
		"withdraw", effectiveWithdrawer,
	)

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeRegisterRevenue,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.DeployerAddress),
				sdk.NewAttribute(types.AttributeKeyContract, msg.ContractAddress),
				sdk.NewAttribute(types.AttributeKeyWithdrawerAddress, effectiveWithdrawer),
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

	contract := sdk.MustAccAddressFromBech32(msg.ContractAddress)

	revenue, found := k.GetRevenue(ctx, contract)
	if !found {
		return nil, sdkerrors.Wrapf(
			types.ErrRevenueContractNotRegistered,
			"contract %s is not registered", msg.ContractAddress,
		)
	}

	// error if the msg deployer address is not the same as the fee's deployer
	if msg.DeployerAddress != revenue.DeployerAddress {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrUnauthorized,
			"%s is not the contract deployer", msg.DeployerAddress,
		)
	}

	// check if updating revenue to default withdrawer
	if msg.WithdrawerAddress == revenue.DeployerAddress {
		msg.WithdrawerAddress = ""
	}

	// revenue with the given withdraw address is already registered
	if msg.WithdrawerAddress == revenue.WithdrawerAddress {
		return nil, sdkerrors.Wrapf(
			types.ErrRevenueAlreadyRegistered,
			"revenue with withdraw address %s", msg.WithdrawerAddress,
		)
	}

	// only delete withdrawer map if is not default
	if revenue.WithdrawerAddress != "" {
		k.DeleteWithdrawerMap(ctx, sdk.MustAccAddressFromBech32(revenue.WithdrawerAddress), contract)
	}

	// only add withdrawer map if new entry is not default
	if msg.WithdrawerAddress != "" {
		k.SetWithdrawerMap(
			ctx,
			sdk.MustAccAddressFromBech32(msg.WithdrawerAddress),
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

	// contract := common.HexToAddress(msg.ContractAddress)
	contract := sdk.MustAccAddressFromBech32(msg.ContractAddress)

	fee, found := k.GetRevenue(ctx, contract)
	if !found {
		return nil, sdkerrors.Wrapf(
			types.ErrRevenueContractNotRegistered,
			"contract %s is not registered", msg.ContractAddress,
		)
	}

	if msg.DeployerAddress != fee.DeployerAddress {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrUnauthorized,
			"%s is not the contract deployer", msg.DeployerAddress,
		)
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
