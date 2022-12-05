package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	authforktypes "github.com/CosmosContracts/juno/v12/x/auth/types"
	feeshare "github.com/CosmosContracts/juno/v12/x/feeshare/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// FeeSharePayoutDecorator Run his after we already deduct the fee from the account with
// the ante.NewDeductFeeDecorator() decorator. We pull funds from the FeeCollector ModuleAccount
type FeeSharePayoutDecorator struct {
	ak             authforktypes.AccountKeeper
	bankKeeper     authforktypes.BankKeeper
	feegrantKeeper authforktypes.FeegrantKeeper
	feesharekeeper authforktypes.FeeShareKeeper
}

func NewFeeSharePayoutDecorator(ak authforktypes.AccountKeeper, bk authforktypes.BankKeeper, fk authforktypes.FeegrantKeeper, fs authforktypes.FeeShareKeeper) FeeSharePayoutDecorator {
	return FeeSharePayoutDecorator{
		ak:             ak,
		bankKeeper:     bk,
		feegrantKeeper: fk,
		feesharekeeper: fs,
	}
}

func (fsd FeeSharePayoutDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	err = FeeSharePayout(ctx, fsd.bankKeeper, feeTx.GetFee(), fsd.feesharekeeper, tx.GetMsgs())
	if err != nil {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}

	return next(ctx, tx, simulate)
}

// FeePayLogic takes the total fees and splits them based on the governance params
// and the number of contracts we are executing on.
// This returns the amount of fees each contract developer should get.
// tested in ante_test.go
func FeePayLogic(fees sdk.Coins, govPercent sdk.Dec, numPairs int) sdk.Coins {
	var splitFees sdk.Coins
	for _, c := range fees {
		amt := c.Amount.QuoRaw(int64(100))                             // makes the fee smaller (500 -> 5)
		amt = amt.MulRaw(int64(govPercent.MulInt64(100).RoundInt64())) // multiple by the govPercent
		amt = amt.QuoRaw(int64(numPairs))                              // split between the pairs evenly

		reward := sdk.NewCoin(c.Denom, amt)
		if !reward.Amount.IsZero() {
			splitFees = append(splitFees, reward)
		}
	}
	return splitFees
}

// FeeSharePayout takes the total fees and redistributes 50% (or param set) to the contract developers
// provided they opted-in to payments.
func FeeSharePayout(ctx sdk.Context, bankKeeper authforktypes.BankKeeper, totalFees sdk.Coins, revKeeper authforktypes.FeeShareKeeper, msgs []sdk.Msg) error {

	params := revKeeper.GetParams(ctx)
	if !params.EnableFeeShare {
		return nil
	}

	// Get only allowed governance fees to be paid (helps for taxes)
	// Juno v13 will have a globalFee for setting more allowed denoms later.
	var fees sdk.Coins
	if len(params.AllowedDenoms) == 0 {
		// If empty, we allow all denoms to be used as payment
		fees = totalFees
	} else {
		for _, fee := range totalFees {
			for _, allowed := range params.AllowedDenoms {
				if fee.Denom == allowed {
					fees = fees.Add(fee)
				}
			}
		}
	}

	// Get valid withdraw addresses from contracts
	toPay := make([]sdk.AccAddress, 0)
	for _, msg := range msgs {
		if _, ok := msg.(*wasmtypes.MsgExecuteContract); ok {
			contractAddr, err := sdk.AccAddressFromBech32(msg.(*wasmtypes.MsgExecuteContract).Contract)
			if err != nil {
				return err
			}

			shareData, _ := revKeeper.GetFeeShare(ctx, contractAddr)

			withdrawAddr := shareData.GetWithdrawerAddr()
			if withdrawAddr != nil && !withdrawAddr.Empty() {
				toPay = append(toPay, withdrawAddr)
			}
		}
	}

	// FeeShare logic payouts for contracts
	numPairs := len(toPay)
	if numPairs > 0 {
		govPercent := params.DeveloperShares
		splitFees := FeePayLogic(fees, govPercent, numPairs)

		// pay fees evenly between all withdraw addresses
		for _, withdrawAddr := range toPay {
			err := bankKeeper.SendCoinsFromModuleToAccount(ctx, types.FeeCollectorName, withdrawAddr, splitFees)
			if err != nil {
				return sdkerrors.Wrapf(feeshare.ErrFeeSharePayment, "failed to pay fees to contract developer: %s", err.Error())
			}
		}
	}

	return nil
}
