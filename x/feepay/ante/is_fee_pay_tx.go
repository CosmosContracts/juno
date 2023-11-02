package ante

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	feepayhelpers "github.com/CosmosContracts/juno/v18/x/feepay/helpers"
	feepaykeeper "github.com/CosmosContracts/juno/v18/x/feepay/keeper"
)

// MsgFilterDecorator defines an AnteHandler decorator that only checks and saves if a
type MsgIsFeePayTx struct {
	feePayKeeper feepaykeeper.Keeper
	isFeePayTx   *bool
}

func NewIsFeePayTxDecorator(feepaykeeper feepaykeeper.Keeper, isFeePayTx *bool) MsgIsFeePayTx {
	return MsgIsFeePayTx{
		feePayKeeper: feepaykeeper,
		isFeePayTx:   isFeePayTx,
	}
}

// AnteHandle performs an AnteHandler check that returns an error if the tx contains a message
// that is blocked.
func (mfd MsgIsFeePayTx) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// Check if a fee tx
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	// Flag a transaction as a fee pay transaction
	*mfd.isFeePayTx = feepayhelpers.IsValidFeePayTransaction(ctx, mfd.feePayKeeper, feeTx)

	// Call next handler
	return next(ctx, tx, simulate)
}
