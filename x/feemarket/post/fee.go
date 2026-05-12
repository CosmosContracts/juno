package post

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/CosmosContracts/juno/v30/app/ante/decorators"
	feemarketkeeper "github.com/CosmosContracts/juno/v30/x/feemarket/keeper"
	feemarkettypes "github.com/CosmosContracts/juno/v30/x/feemarket/types"
)

// BankSendGasConsumption is the gas consumption of the bank sends that occur during feemarket handler execution.
const BankSendGasConsumption = 12490

// FeeMarketDeductDecorator deducts fees from the fee payer based off of the current state of the feemarket.
// The fee payer is the fee granter (if specified) or first signer of the tx.
// If the fee payer does not have the funds to pay for the fees, return an InsufficientFunds error.
// If there is an excess between the given fee and the on-chain min base fee is given as a tip.
// Call next PostHandler if fees successfully deducted.
// CONTRACT: Tx must implement FeeTx interface
type FeeMarketDeductDecorator struct {
	accountKeeper   authkeeper.AccountKeeper
	bankKeeper      bankkeeper.Keeper
	feemarketKeeper feemarketkeeper.Keeper
}

func NewFeeMarketDeductDecorator(ak authkeeper.AccountKeeper, bk bankkeeper.Keeper, fmk feemarketkeeper.Keeper) FeeMarketDeductDecorator {
	return FeeMarketDeductDecorator{
		accountKeeper:   ak,
		bankKeeper:      bk,
		feemarketKeeper: fmk,
	}
}

// PostHandle deducts the fee from the fee payer based on the min base fee and the gas consumed in the gasmeter.
// If there is a difference between the provided fee and the min-base fee, the difference is paid as a tip.
// Fees are sent to the x/feemarket fee-collector address.
func (dfd FeeMarketDeductDecorator) PostHandle(ctx sdk.Context, tx sdk.Tx, simulate, success bool, next sdk.PostHandler) (sdk.Context, error) {
	// GenTx consume no fee
	if ctx.BlockHeight() == 0 {
		return next(ctx, tx, simulate, success)
	}

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	if !simulate && ctx.BlockHeight() > 0 && feeTx.GetGas() == 0 {
		return ctx, errorsmod.Wrap(sdkerrors.ErrInvalidGasLimit, "must provide positive gas")
	}

	// update fee market params
	params, err := dfd.feemarketKeeper.GetParams(ctx)
	if err != nil {
		return ctx, errorsmod.Wrapf(err, "unable to get fee market params")
	}

	// return if disabled
	if !params.Enabled {
		return next(ctx, tx, simulate, success)
	}

	enabledHeight, err := dfd.feemarketKeeper.GetEnabledHeight(ctx)
	if err != nil {
		return ctx, errorsmod.Wrapf(err, "unable to get fee market enabled height")
	}

	// if the current height is that which enabled the feemarket or lower, skip deduction
	if ctx.BlockHeight() <= enabledHeight {
		return next(ctx, tx, simulate, success)
	}

	// update fee market state
	state, err := dfd.feemarketKeeper.GetState(ctx)
	if err != nil {
		return ctx, errorsmod.Wrapf(err, "unable to get fee market state")
	}

	feeCoins := feeTx.GetFee()
	gas := ctx.GasMeter().GasConsumed() // use context gas consumed

	if len(feeCoins) > 1 {
		return ctx, errorsmod.Wrapf(feemarkettypes.ErrTooManyFeeCoins, "got length %d", len(feeCoins))
	}

	// if simulating and user did not provider a fee - create a dummy value for them
	var (
		tip     = sdk.NewCoin(params.FeeDenom, math.ZeroInt())
		payCoin = sdk.NewCoin(params.FeeDenom, math.ZeroInt())
	)
	if !simulate {
		switch {
		case len(feeCoins) > 0:
			payCoin = feeCoins[0]
		default:
			// Feepay path: the user submitted --fees 0 and x/feepay has
			// already deposited the required fee into feemarket-fee-collector
			// in the ante handler. Treat the module's current balance in the
			// fee denom as the effective fee for accounting (CheckTxFee will
			// then split it into consumedFee + tip).
			moduleAddr := dfd.accountKeeper.GetModuleAddress(feemarkettypes.FeeCollectorName)
			payCoin = dfd.bankKeeper.GetBalance(ctx, moduleAddr, params.FeeDenom)
			if payCoin.IsZero() {
				return ctx, errorsmod.Wrapf(feemarkettypes.ErrNoFeeCoins, "got length %d and feemarket-fee-collector empty", len(feeCoins))
			}
		}
	}

	feeGas := int64(feeTx.GetGas())

	currentGasPrice, err := dfd.feemarketKeeper.GetCurrentGasPrice(ctx, payCoin.GetDenom())
	if err != nil {
		return ctx, errorsmod.Wrapf(err, "unable to get min gas price for denom %s", payCoin.GetDenom())
	}

	ctx.Logger().Debug("fee deduct post handle",
		"gas prices", currentGasPrice,
		"gas consumed", gas,
	)

	if !simulate {
		payCoin, tip, err = decorators.CheckTxFee(ctx, currentGasPrice, payCoin, feeGas, false)
		if err != nil {
			return ctx, err
		}
	}

	ctx.Logger().Debug("fee deduct post handle",
		"fee", payCoin,
		"tip", tip,
	)

	if err := dfd.PayOutFeeAndTip(ctx, payCoin, tip); err != nil {
		return ctx, err
	}

	err = state.Update(gas, params)
	if err != nil {
		return ctx, errorsmod.Wrapf(err, "unable to update fee market state")
	}

	err = dfd.feemarketKeeper.SetState(ctx, state)
	if err != nil {
		return ctx, errorsmod.Wrapf(err, "unable to set fee market state")
	}

	if simulate {
		// consume the gas that would be consumed during normal execution
		ctx.GasMeter().ConsumeGas(BankSendGasConsumption, "simulation send gas consumption")
	}

	return next(ctx, tx, simulate, success)
}

// PayOutFeeAndTip deducts the provided fee and tip from the fee payer.
// If the tx uses a feegranter, the fee granter address will pay the fee instead of the tx signer.
func (dfd FeeMarketDeductDecorator) PayOutFeeAndTip(ctx sdk.Context, fee, tip sdk.Coin) error {
	params, err := dfd.feemarketKeeper.GetParams(ctx)
	if err != nil {
		return errorsmod.Wrapf(err, "error getting feemarket params")
	}

	// Cap fee + tip at the current feemarket-fee-collector balance. The
	// x/feeshare ante decorator may have already drained the dev's share
	// from the same module account in this same tx, leaving less than the
	// originally-paid fee available here. Without capping, drains would fail
	// with "insufficient funds" and roll the whole tx (including the message
	// effects) back. Cap the fee first; if there is balance left over, the
	// rest goes to the proposer as tip.
	if !fee.IsNil() && fee.Denom != "" {
		moduleAddr := dfd.accountKeeper.GetModuleAddress(feemarkettypes.FeeCollectorName)
		moduleBal := dfd.bankKeeper.GetBalance(ctx, moduleAddr, fee.Denom).Amount
		if fee.Amount.GT(moduleBal) {
			fee.Amount = moduleBal
		}
		remaining := moduleBal.Sub(fee.Amount)
		if !tip.IsNil() && tip.Denom == fee.Denom && tip.Amount.GT(remaining) {
			tip.Amount = remaining
		}
	}

	var events sdk.Events

	// deduct the fees and tip
	if !fee.IsNil() && !fee.IsZero() {
		err := DeductCoins(dfd.bankKeeper, ctx, sdk.NewCoins(fee), params.DistributeFees)
		if err != nil {
			return err
		}

		events = append(events, sdk.NewEvent(
			feemarkettypes.EventTypeFeePay,
			sdk.NewAttribute(sdk.AttributeKeyFee, fee.String()),
		))
	}

	proposer := sdk.AccAddress(ctx.BlockHeader().ProposerAddress)
	if !tip.IsNil() && !tip.IsZero() {
		err := SendTip(dfd.bankKeeper, ctx, proposer, sdk.NewCoins(tip))
		if err != nil {
			return err
		}

		events = append(events, sdk.NewEvent(
			feemarkettypes.EventTypeTipPay,
			sdk.NewAttribute(feemarkettypes.AttributeKeyTip, tip.String()),
			sdk.NewAttribute(feemarkettypes.AttributeKeyTipPayee, proposer.String()),
		))
	}

	ctx.EventManager().EmitEvents(events)
	return nil
}

// DeductCoins deducts coins from the given account.
// Coins can be sent to the default fee collector (
// causes coins to be distributed to stakers) or kept in the fee collector account (soft burn).
func DeductCoins(bankKeeper bankkeeper.Keeper, ctx sdk.Context, coins sdk.Coins, distributeFees bool) error {
	if distributeFees {
		err := bankKeeper.SendCoinsFromModuleToModule(ctx, feemarkettypes.FeeCollectorName, authtypes.FeeCollectorName, coins)
		if err != nil {
			return err
		}
	}
	return nil
}

// SendTip sends a tip to the current block proposer.
func SendTip(bankKeeper bankkeeper.Keeper, ctx sdk.Context, proposer sdk.AccAddress, coins sdk.Coins) error {
	err := bankKeeper.SendCoinsFromModuleToAccount(ctx, feemarkettypes.FeeCollectorName, proposer, coins)
	if err != nil {
		return err
	}

	return nil
}
