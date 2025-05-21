package ante

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	feemarketante "github.com/CosmosContracts/juno/v30/x/feemarket/ante"
	feepayhelpers "github.com/CosmosContracts/juno/v30/x/feepay/helpers"
	feepaykeeper "github.com/CosmosContracts/juno/v30/x/feepay/keeper"
)

// MsgIsFeePayTx defines an AnteHandler decorator that only checks and saves if a
type MsgIsFeePayTx struct {
	feePayKeeper       feepaykeeper.Keeper
	feePayDecorator    *DeductFeeDecorator
	feeMarketDecorator *feemarketante.FeeMarketCheckDecorator
	isFeePayTx         *bool
}

func NewFeeRouteDecorator(feePayKeeper feepaykeeper.Keeper, feePayDecorator *DeductFeeDecorator, feeMarketDecorator *feemarketante.FeeMarketCheckDecorator, isFeePayTx *bool) MsgIsFeePayTx {
	return MsgIsFeePayTx{
		feePayKeeper:       feePayKeeper,
		feePayDecorator:    feePayDecorator,
		feeMarketDecorator: feeMarketDecorator,
		isFeePayTx:         isFeePayTx,
	}
}

// This empty ante is used to call AnteHandles that are not attached
// to the main AnteHandler.
var (
	EmptyAnte = func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
		return ctx, nil
	}
)

// This handle is responsible for routing the transaction to the fee decorators
// in the right order.
func (mfd MsgIsFeePayTx) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// Check if a fee tx
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	// Flag a transaction as a fee pay transaction
	*mfd.isFeePayTx = feepayhelpers.IsValidFeePayTransaction(ctx, mfd.feePayKeeper, feeTx)

	// If a FeePayTx, call FeePay decorator then feemarket decorator.
	// Otherwise, call feemarket decorator then FeePay decorator.
	//
	// This logic is necessary in the case the FeePay decorator fails,
	// the feemarket decorator will still be called to handle fees.
	if *mfd.isFeePayTx {
		if ctx, err := mfd.feePayDecorator.AnteHandle(ctx, tx, simulate, EmptyAnte); err != nil {
			return ctx, err
		}

		if ctx, err := mfd.feeMarketDecorator.AnteHandle(ctx, tx, simulate, EmptyAnte); err != nil {
			return ctx, err
		}
	} else {
		if ctx, err := mfd.feeMarketDecorator.AnteHandle(ctx, tx, simulate, EmptyAnte); err != nil {
			return ctx, err
		}

		if ctx, err := mfd.feePayDecorator.AnteHandle(ctx, tx, simulate, EmptyAnte); err != nil {
			return ctx, err
		}
	}

	// Call next handler
	return next(ctx, tx, simulate)
}
