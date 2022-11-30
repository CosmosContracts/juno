package keeper

import (
	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/CosmosContracts/juno/v12/x/feeshare/types"
)

// TODO: https://github.com/cosmos/cosmos-sdk/blob/main/x/slashing/keeper/hooks.go

var _ types.RevenueHooks = Hooks{}

// Hooks wrapper struct for fees keeper
type Hooks struct {
	k Keeper
}

// Hooks return the wrapper hooks struct for the Keeper
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// PostTxProcessing is a wrapper for calling the EVM PostTxProcessing hook on
// the module keeper
// func (h Hooks) PostTxProcessing(ctx sdk.Context, msg core.Message) error {
func (h Hooks) PostTxProcessing(ctx sdk.Context, msg wasmtypes.MsgExecuteContract, feeTx sdk.FeeTx) error {
	return h.k.PostTxProcessing(ctx, msg, feeTx)
}

// PostTxProcessing implements RevenueHooks.PostTxProcessing. After each successful
// interaction with a registered contract, the contract deployer (or, if set,
// the withdraw address) receives a share from the transaction fees paid by the
// transaction sender.
func (k Keeper) PostTxProcessing(
	ctx sdk.Context,
	msg wasmtypes.MsgExecuteContract,
	feeTx sdk.FeeTx,
) error {
	// check if the fees are globally enabled
	params := k.GetParams(ctx)
	if !params.EnableRevenue {
		return nil
	}

	if len(msg.Contract) == 0 {
		return nil
	}

	contract, err := sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return err
	}

	// if the contract is not registered to receive fees, do nothing
	feeSplit, found := k.GetRevenue(ctx, contract)
	if !found {
		return nil
	}

	withdrawer := feeSplit.GetWithdrawerAddr()
	if len(withdrawer) == 0 {
		withdrawer = feeSplit.GetDeployerAddr()
	}

	// TODO: What if the fee is not in juno? add logic to loop through coins fee paid in
	denom := "ujuno"

	// txFee := sdk.NewIntFromUint64(receipt.GasUsed).Mul(sdk.NewIntFromBigInt(msg.GasPrice()))
	// txFee := sdk.NewIntFromUint64(ctx.BlockGasMeter().GasConsumed()).Mul(sdk.NewIntFromBigInt(msg.GasPrice()))
	txFee := sdk.NewIntFromUint64(ctx.BlockGasMeter().GasConsumed()).Mul(sdk.NewIntFromBigInt(feeTx.GetFee().AmountOf(denom).BigInt()))
	developerFee := txFee.ToDec().Mul(params.DeveloperShares).TruncateInt() // params.DeveloperShares = 50% by default

	fees := sdk.Coins{{Denom: denom, Amount: developerFee}}

	// distribute the fees to the contract deployer / withdraw address
	err = k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		k.feeCollectorName,
		withdrawer,
		fees,
	)
	if err != nil {
		return sdkerrors.Wrapf(
			err,
			"fee collector account failed to distribute developer fees (%s) to withdraw address %s. contract %s",
			fees, withdrawer, contract,
		)
	}

	defer func() {
		if developerFee.IsInt64() {
			telemetry.IncrCounterWithLabels(
				[]string{types.ModuleName, "distribute", "total"},
				float32(developerFee.Int64()),
				[]metrics.Label{
					telemetry.NewLabel("sender", msg.Sender),
					telemetry.NewLabel("withdraw_address", withdrawer.String()),
					telemetry.NewLabel("contract", feeSplit.ContractAddress),
				},
			)
		}
	}()

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeDistributeDevRevenue,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
				sdk.NewAttribute(types.AttributeKeyContract, contract.String()),
				sdk.NewAttribute(types.AttributeKeyWithdrawerAddress, withdrawer.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, developerFee.String()),
			),
		},
	)

	return nil
}
