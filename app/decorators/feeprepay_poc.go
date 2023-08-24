package decorators

import (
	globalfeekeeper "github.com/CosmosContracts/juno/v17/x/globalfee/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

type MsgFeePrepayDecorator struct {
	BankKeeper bankkeeper.Keeper
	// TODO: This could be incorrect. We may need the full AuthKeeper/AccountKeeper, not just from the ante
	AccountKeeper   authante.AccountKeeper
	GlobalFeeKeeper globalfeekeeper.Keeper
}

func NewMsgFeePrepayDecorator(bank bankkeeper.Keeper, auth authante.AccountKeeper, globalFee globalfeekeeper.Keeper) MsgFeePrepayDecorator {
	return MsgFeePrepayDecorator{
		BankKeeper:      bank,
		AccountKeeper:   auth,
		GlobalFeeKeeper: globalFee,
	}
}

func (fpd MsgFeePrepayDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// if tx is read only
	// tx := new sdk.Tx(...tx, set the new fee)

	// This may not be a FeeTx, so if error then don't exit early
	// feeTx, ok := tx.(sdk.FeeTx)
	// if !ok {
	// 	return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must implement the sdk.FeeTx interface")
	// }

	// get gas

	// p := fpd.GlobalFeeKeeper.GetParams(ctx)
	// var minGasprice sdk.DecCoins
	// for _, c := range p.MinimumGasPrices {
	// 	if c.Denom == "ujuno" {
	// 		// get that amount
	// 		amt := c.Amount
	// 	}
	// }

	// TODOL: Thuis could give you funds and then make your next Tx successful.
	// fpd.BankKeeper.MintCoins(ctx, "bank", coinsAmt)
	// fpd.BankKeeper.SendCoinsFromModuleToAccount(ctx, "bank", "userAccount", coinsAmt)
	// auto feeprepay with the TxFee or somethuing here? as an option
	// set the Tx fee to be correct, set the accoiunt to be new and work, and then continue on

	return next(ctx, tx, simulate)
}

// TODO: Future: execute contract only.
// func hasInvalidExecuteMsgs(msgs []sdk.Msg) bool {
// 	for _, msg := range msgs {
// 		if _, ok := msg.(*wasmtypes.msgExecuteContract); ok {
// 			return true
// 		}
// 	}

// 	return false
// }
