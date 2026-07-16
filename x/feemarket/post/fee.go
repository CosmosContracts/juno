package post

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

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
	feepayhelpers "github.com/CosmosContracts/juno/v30/x/feepay/helpers"
	feepaykeeper "github.com/CosmosContracts/juno/v30/x/feepay/keeper"
	feepaytypes "github.com/CosmosContracts/juno/v30/x/feepay/types"
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
	feepayKeeper    feepaykeeper.Keeper
	stakingKeeper   feemarkettypes.StakingKeeper
}

func NewFeeMarketDeductDecorator(ak authkeeper.AccountKeeper, bk bankkeeper.Keeper, fmk feemarketkeeper.Keeper, fpk feepaykeeper.Keeper, sk feemarkettypes.StakingKeeper) FeeMarketDeductDecorator {
	return FeeMarketDeductDecorator{
		accountKeeper:   ak,
		bankKeeper:      bk,
		feemarketKeeper: fmk,
		feepayKeeper:    fpk,
		stakingKeeper:   sk,
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
		tip        = sdk.NewCoin(params.FeeDenom, math.ZeroInt())
		payCoin    = sdk.NewCoin(params.FeeDenom, math.ZeroInt())
		isFeePayTx = false
		feeGas     = int64(feeTx.GetGas())
		skipDeduct = false
	)
	if !simulate {
		switch {
		case len(feeCoins) > 0:
			payCoin = feeCoins[0]
		default:
			// Zero-fee tx. Either a feepay tx (the ante escrowed exactly
			// price × gasLimit out of the contract's feepay balance into
			// feemarket-fee-collector) or a bypass-min-fee tx (IBC relayer
			// messages; the ante escrowed nothing).
			isFeePayTx = feepayhelpers.IsValidFeePayTransaction(ctx, dfd.feepayKeeper, feeTx)
			if !isFeePayTx {
				// Bypass tx: nothing escrowed, nothing to deduct. Still record
				// the gas consumed in the fee market state below.
				skipDeduct = true
			}
		}
	}

	if !skipDeduct {
		currentGasPrice, err := dfd.feemarketKeeper.GetCurrentGasPrice(ctx, payCoin.GetDenom())
		if err != nil {
			return ctx, errorsmod.Wrapf(err, "unable to get min gas price for denom %s", payCoin.GetDenom())
		}

		if isFeePayTx {
			// Bound the effective fee to exactly THIS tx's escrow:
			// ceil(gasPrice × gasLimit), mirroring the ante's handleZeroFees.
			// Never read the whole collector balance — with DistributeFees =
			// false the module account accumulates funds across txs, and a
			// feepay tx must not be able to claim them.
			escrowAmount := currentGasPrice.Amount.Mul(math.LegacyNewDec(feeGas)).Ceil().RoundInt()
			payCoin = sdk.NewCoin(currentGasPrice.Denom, escrowAmount)

			// Defense in depth: never exceed what is actually in the module
			// account. (The ante deposited exactly escrowAmount this tx.)
			moduleAddr := dfd.accountKeeper.GetModuleAddress(feemarkettypes.FeeCollectorName)
			balance := dfd.bankKeeper.GetBalance(ctx, moduleAddr, payCoin.Denom)
			if payCoin.Amount.GT(balance.Amount) {
				payCoin.Amount = balance.Amount
			}
			if payCoin.IsZero() {
				return ctx, errorsmod.Wrapf(feemarkettypes.ErrNoFeeCoins, "feepay escrow missing from feemarket-fee-collector")
			}
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

		if isFeePayTx {
			// Feepay txs generate NO proposer tip: the "tip" here is only the
			// unused-gas remainder of the contract's escrow (the contract was
			// charged for the full gas limit up front). Paying it to the
			// proposer would let proposers drain funded feepay contracts with
			// gas-padded txs. Refund it to the contract's feepay balance.
			if err := dfd.PayOutFeeAndRefundFeePay(ctx, feeTx, payCoin, tip); err != nil {
				return ctx, err
			}
		} else {
			if err := dfd.PayOutFeeAndTip(ctx, payCoin, tip); err != nil {
				return ctx, err
			}
		}
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
	fee, tip = dfd.capAtCollectorBalance(ctx, fee, tip)

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

	if !tip.IsNil() && !tip.IsZero() {
		// Resolve the proposer's CONSENSUS address to the validator operator
		// account. ProposerAddress is a consensus (ed25519) address — casting
		// it straight to an AccAddress produces an account no operator key
		// controls, stranding the tip forever.
		proposer, found := dfd.proposerOperatorAccount(ctx)
		if !found {
			// No operator account resolvable (should not happen for a block
			// proposer). Leave the tip in the fee collector instead of
			// stranding it at an unspendable address; it is distributed (or
			// soft-burned) with the next DistributeFees sweep.
			ctx.Logger().Error("feemarket post handler: could not resolve proposer operator account; leaving tip in fee collector",
				"proposer_cons_address", sdk.ConsAddress(ctx.BlockHeader().ProposerAddress).String(),
			)
		} else {
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
	}

	ctx.EventManager().EmitEvents(events)
	return nil
}

// PayOutFeeAndRefundFeePay handles the payout for a feepay-covered tx:
//   - the consumed fee is deducted as usual (distributed or soft-burned per
//     params.DistributeFees);
//   - the unused-gas remainder of the escrow is refunded from the
//     feemarket-fee-collector back to the x/feepay module account and
//     re-credited to the executing contract's feepay balance.
//
// No proposer tip is generated for feepay txs.
func (dfd FeeMarketDeductDecorator) PayOutFeeAndRefundFeePay(ctx sdk.Context, feeTx sdk.FeeTx, fee, refund sdk.Coin) error {
	params, err := dfd.feemarketKeeper.GetParams(ctx)
	if err != nil {
		return errorsmod.Wrapf(err, "error getting feemarket params")
	}

	// cap at the collector balance (x/feeshare pays nothing for feepay txs —
	// the tx fee is zero — but stay defensive)
	fee, refund = dfd.capAtCollectorBalance(ctx, fee, refund)

	var events sdk.Events

	if !fee.IsNil() && !fee.IsZero() {
		if err := DeductCoins(dfd.bankKeeper, ctx, sdk.NewCoins(fee), params.DistributeFees); err != nil {
			return err
		}

		events = append(events, sdk.NewEvent(
			feemarkettypes.EventTypeFeePay,
			sdk.NewAttribute(sdk.AttributeKeyFee, fee.String()),
		))
	}

	if !refund.IsNil() && !refund.IsZero() {
		// CONTRACT: a valid feepay tx has exactly one MsgExecuteContract on a
		// registered contract (enforced by IsValidFeePayTransaction).
		msgs := feeTx.GetMsgs()
		if len(msgs) != 1 {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "feepay tx must contain exactly one message, got %d", len(msgs))
		}
		cw, ok := msgs[0].(*wasmtypes.MsgExecuteContract)
		if !ok {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "feepay tx message must be a MsgExecuteContract, got %T", msgs[0])
		}

		contract, err := dfd.feepayKeeper.GetContract(ctx, cw.GetContract())
		if err != nil {
			return errorsmod.Wrapf(err, "error getting feepay contract %s for escrow refund", cw.GetContract())
		}

		if err := dfd.bankKeeper.SendCoinsFromModuleToModule(ctx, feemarkettypes.FeeCollectorName, feepaytypes.ModuleName, sdk.NewCoins(refund)); err != nil {
			return errorsmod.Wrapf(err, "error refunding feepay escrow")
		}

		dfd.feepayKeeper.SetContractBalance(ctx, contract, contract.Balance+refund.Amount.Uint64())

		events = append(events, sdk.NewEvent(
			feemarkettypes.EventTypeFeePayRefund,
			sdk.NewAttribute(feemarkettypes.AttributeKeyRefund, refund.String()),
			sdk.NewAttribute(feemarkettypes.AttributeKeyRefundPayee, cw.GetContract()),
		))
	}

	ctx.EventManager().EmitEvents(events)
	return nil
}

// capAtCollectorBalance caps fee + remainder at the feemarket-fee-collector's
// current balance: fee first, then the remainder gets what is left.
func (dfd FeeMarketDeductDecorator) capAtCollectorBalance(ctx sdk.Context, fee, remainder sdk.Coin) (sdk.Coin, sdk.Coin) {
	if fee.IsNil() || fee.Denom == "" {
		return fee, remainder
	}

	moduleAddr := dfd.accountKeeper.GetModuleAddress(feemarkettypes.FeeCollectorName)
	moduleBal := dfd.bankKeeper.GetBalance(ctx, moduleAddr, fee.Denom).Amount
	if fee.Amount.GT(moduleBal) {
		fee.Amount = moduleBal
	}
	remaining := moduleBal.Sub(fee.Amount)
	if !remainder.IsNil() && remainder.Denom == fee.Denom && remainder.Amount.GT(remaining) {
		remainder.Amount = remaining
	}

	return fee, remainder
}

// proposerOperatorAccount resolves the current block proposer's consensus
// address to the validator operator's account address, mirroring
// x/distribution's proposer attribution.
func (dfd FeeMarketDeductDecorator) proposerOperatorAccount(ctx sdk.Context) (sdk.AccAddress, bool) {
	consAddr := sdk.ConsAddress(ctx.BlockHeader().ProposerAddress)
	if len(consAddr) == 0 {
		return nil, false
	}

	validator, err := dfd.stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	if err != nil {
		return nil, false
	}

	valAddr, err := sdk.ValAddressFromBech32(validator.GetOperator())
	if err != nil {
		return nil, false
	}

	return sdk.AccAddress(valAddr), true
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
