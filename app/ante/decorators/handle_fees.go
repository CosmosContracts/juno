package decorators

import (
	"bytes"
	"math"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	feemarketkeeper "github.com/CosmosContracts/juno/v30/x/feemarket/keeper"
	feemarkettypes "github.com/CosmosContracts/juno/v30/x/feemarket/types"
	feepayhelpers "github.com/CosmosContracts/juno/v30/x/feepay/helpers"
	feepaykeeper "github.com/CosmosContracts/juno/v30/x/feepay/keeper"
	feepaytypes "github.com/CosmosContracts/juno/v30/x/feepay/types"
)

const (
	// gasPricePrecisionMultiplier scales normalized gas prices to integer
	// priorities (10^6, i.e. 6 digits of precision), precomputed as an
	// integer so the priority path never touches float math.
	gasPricePrecisionMultiplier = int64(1_000_000)

	// MaxBypassMinFeeMsgGasUsage is the maximum gas limit a zero-fee tx made up
	// exclusively of bypass message types (IBC relayer messages) may request.
	// It bounds the free gas a relayer can consume per tx; anything heavier
	// must pay fees like everyone else. Matches the pre-v30 x/globalfee value.
	MaxBypassMinFeeMsgGasUsage = uint64(2_000_000)
)

type DeductFeeDecorator struct {
	feemarketkeeper feemarketkeeper.Keeper

	innerDecorator    InnerDeductFeeDecorator
	fallbackDecorator sdk.AnteDecorator
}

func NewDeductFeeDecorator(fpk feepaykeeper.Keeper, fmk feemarketkeeper.Keeper, ak authkeeper.AccountKeeper, bk bankkeeper.Keeper, fgk feegrantkeeper.Keeper, bondDenom string, bypassMinFeeMsgTypes []string, fallbackDecorator sdk.AnteDecorator) DeductFeeDecorator {
	return DeductFeeDecorator{
		feemarketkeeper: fmk,
		innerDecorator: newInnerDeductFeeDecorator(
			fpk, fmk, ak, bk, fgk, bondDenom, bypassMinFeeMsgTypes,
		),
		fallbackDecorator: fallbackDecorator,
	}
}

// InnerDeductFeeDecorator deducts fees from the first signer of the tx
// If the first signer does not have the funds to pay for the fees, return with InsufficientFunds error
// Call next AnteHandler if fees successfully deducted
// CONTRACT: Tx must implement FeeTx interface to use DeductFeeDecorator
//
// Additionally, the Deduct Fee ante is a fork of the SDK's DeductFeeDecorator. This decorator looks for single
// message transactions with no provided fee. If they correspond to a registered FeePay Contract, the FeePay
// module will cover the cost of the fee (if the balance permits).
type InnerDeductFeeDecorator struct {
	feepayKeeper         feepaykeeper.Keeper
	feemarketKeeper      feemarketkeeper.Keeper
	accountKeeper        authkeeper.AccountKeeper
	bankKeeper           bankkeeper.Keeper
	feegrantKeeper       feegrantkeeper.Keeper
	bondDenom            string
	bypassMinFeeMsgTypes []string
}

func newInnerDeductFeeDecorator(fpk feepaykeeper.Keeper, fmk feemarketkeeper.Keeper, ak authkeeper.AccountKeeper, bk bankkeeper.Keeper, fgk feegrantkeeper.Keeper, bondDenom string, bypassMinFeeMsgTypes []string) InnerDeductFeeDecorator {
	return InnerDeductFeeDecorator{
		feepayKeeper:         fpk,
		feemarketKeeper:      fmk,
		accountKeeper:        ak,
		bankKeeper:           bk,
		feegrantKeeper:       fgk,
		bondDenom:            bondDenom,
		bypassMinFeeMsgTypes: bypassMinFeeMsgTypes,
	}
}

// AnteHandle calls the feemarket antehandler if the keeper is enabled.  If disabled, the fallback
// fee antehandler is fallen back to.
func (dfd DeductFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	params, err := dfd.feemarketkeeper.GetParams(ctx)
	if err != nil {
		return ctx, err
	}
	if params.Enabled {
		return dfd.innerDecorator.anteHandle(ctx, tx, simulate, next)
	}
	if dfd.fallbackDecorator != nil {
		return dfd.fallbackDecorator.AnteHandle(ctx, tx, simulate, next)
	}

	return next(ctx, tx, simulate)
}

func (dfd InnerDeductFeeDecorator) HandleFees(ctx sdk.Context, feeTx sdk.FeeTx, fee sdk.Coin, isValidFeepayTx bool) error {
	if addr := dfd.accountKeeper.GetModuleAddress(authtypes.FeeCollectorName); addr == nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "fee collector module account (%s) has not been set", authtypes.FeeCollectorName)
	}

	feePayer := feeTx.FeePayer()
	feeGranter := feeTx.FeeGranter()
	deductFeesFrom := feePayer

	// if feegranter set deduct fee from feegranter account.
	// this works with only when x/feegrant is enabled and the fee granter allows the fee payer to cover their fees.
	if feeGranter != nil {
		feeGranterAddr := sdk.AccAddress(feeGranter)
		feePayerAddr := sdk.AccAddress(feePayer)
		if !bytes.Equal(feeGranterAddr, feePayerAddr) {
			err := dfd.feegrantKeeper.UseGrantedFees(ctx, feeGranterAddr, feePayerAddr, sdk.NewCoins(fee), feeTx.GetMsgs())
			if err != nil {
				return errorsmod.Wrapf(err, "%s does not allow to pay fees for %s", feeGranterAddr, feePayerAddr)
			}
		}

		deductFeesFrom = feeGranter
	}

	deductFeesFromAddr := sdk.AccAddress(deductFeesFrom)

	deductFeesFromAcc := dfd.accountKeeper.GetAccount(ctx, deductFeesFromAddr)
	if deductFeesFromAcc == nil {
		return sdkerrors.ErrUnknownAddress.Wrapf("fee payer address: %s does not exist", deductFeesFromAddr)
	}

	// Define errors per route
	var feePayErr error
	var sdkErr error

	// First try to handle FeePay transactions, if error, try the feemarket route.
	// If not a FeePay transaction, default to the feemarket route.
	if isValidFeepayTx {
		feePayErr = dfd.handleZeroFees(ctx, deductFeesFromAcc, feeTx)
		if feePayErr != nil {
			// Only fall back to user-paid escrow when there is an actual fee to
			// escrow. For a valid feepay tx the user submits --fees 0, so `fee`
			// is a zero coin and sdk.NewCoins(fee) is empty: escrow becomes a
			// no-op and the tx slips past the ante. The msg then runs, the
			// post-handler can't find a fee in feemarket-fee-collector, and
			// the tx fails with sequence already incremented — leaving the
			// user's nonce desynced from their on-chain state. Reject in
			// ante so sequence is not consumed.
			if fee.IsZero() {
				return errorsmod.Wrapf(feePayErr, "feepay cannot cover this tx and no user fee was provided")
			}
			sdkErr = dfd.escrow(ctx, deductFeesFromAcc, sdk.NewCoins(fee))
		}
	} else if !fee.IsZero() {
		// Std sdk route
		sdkErr = dfd.escrow(ctx, deductFeesFromAcc, sdk.NewCoins(fee))
	}

	// If no fee pay error exists, the tx processed successfully. If
	// a sdk error is present, return all errors.
	if sdkErr != nil {
		if feePayErr != nil {
			return errorsmod.Wrapf(feepaytypes.ErrDeductFees, "error deducting fees; fee pay error: %s, sdk error: %s", feePayErr, sdkErr)
		}
		return sdkErr
	}

	events := sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeTx,
			sdk.NewAttribute(sdk.AttributeKeyFee, fee.String()),
			sdk.NewAttribute(sdk.AttributeKeyFeePayer, deductFeesFromAcc.GetAddress().String()),
		),
	}
	ctx.EventManager().EmitEvents(events)

	return nil
}

func (dfd InnerDeductFeeDecorator) anteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// GenTx consume no fee
	if ctx.BlockHeight() == 0 {
		return next(ctx, tx, simulate)
	}

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	isValidFeepayTx := feepayhelpers.IsValidFeePayTransaction(ctx, dfd.feepayKeeper, feeTx)

	if !simulate && ctx.BlockHeight() > 0 && feeTx.GetGas() == 0 {
		return ctx, errorsmod.Wrap(sdkerrors.ErrInvalidGasLimit, "must provide positive gas")
	}

	feeCoins := feeTx.GetFee()

	if !isValidFeepayTx {
		if len(feeCoins) == 0 && !simulate {
			// Min-fee bypass: a zero-fee tx whose messages are ALL in the
			// bypass allow-list (IBC relayer messages) is let through without
			// fee escrow, bounded by a gas ceiling so it cannot be abused for
			// free compute. The post handler skips fee deduction for zero-fee
			// non-feepay txs, so nothing downstream expects an escrow.
			if dfd.isBypassMinFeeTx(tx) {
				if feeTx.GetGas() > MaxBypassMinFeeMsgGasUsage {
					return ctx, errorsmod.Wrapf(sdkerrors.ErrInvalidGasLimit,
						"bypass-min-fee tx gas limit %d exceeds maximum of %d", feeTx.GetGas(), MaxBypassMinFeeMsgGasUsage)
				}
				return next(ctx, tx, simulate)
			}
			return ctx, errorsmod.Wrapf(feemarkettypes.ErrNoFeeCoins, "got length %d", len(feeCoins))
		}
	}

	if len(feeCoins) > 1 {
		return ctx, errorsmod.Wrapf(feemarkettypes.ErrTooManyFeeCoins, "got length %d", len(feeCoins))
	}

	params, err := dfd.feemarketKeeper.GetParams(ctx)
	if err != nil {
		return ctx, errorsmod.Wrapf(err, "unable to get fee market params")
	}

	// Default payCoin to a zero coin in the fee market's fee denom. For a
	// valid feepay tx the user submits with --fees 0 (sdk.ParseCoinsNormalized
	// strips the zero, so feeCoins is empty), and the actual fee is covered by
	// x/feepay in HandleFees — there is no feeCoins[0] to read. Pre-fix this
	// indexed past the end of feeCoins and panicked in CheckTx.
	payCoin := sdk.NewCoin(params.FeeDenom, sdkmath.ZeroInt())
	if !simulate && len(feeCoins) > 0 {
		payCoin = feeCoins[0]
	}

	gas := feeTx.GetGas()
	feeGasPrice, err := dfd.feemarketKeeper.GetCurrentGasPrice(ctx, payCoin.GetDenom())
	if err != nil {
		return ctx, errorsmod.Wrapf(err, "unable to get min gas price for denom %s", payCoin.GetDenom())
	}

	ctx.Logger().Debug("fee deduct ante handle",
		"current gas price", feeGasPrice,
		"fee", feeCoins,
		"gas limit", gas,
	)

	ctx = ctx.WithMinGasPrices(sdk.NewDecCoins(feeGasPrice))

	// CheckTxFee compares the user-provided fee against requiredFee
	// (gasLimit * feeGasPrice). Skip it for valid feepay txs: the user
	// provided no fee on purpose and x/feepay covers the requiredFee from
	// the contract balance in HandleFees below.
	if !simulate && !isValidFeepayTx {
		_, _, checkErr := CheckTxFee(ctx, feeGasPrice, payCoin, int64(gas), true)
		if checkErr != nil {
			return ctx, errorsmod.Wrapf(checkErr, "error checking fee")
		}
	}

	// handle the entire tx fee process
	err = dfd.HandleFees(ctx, feeTx, payCoin, isValidFeepayTx)
	if err != nil {
		return ctx, errorsmod.Wrapf(err, "error escrowing funds")
	}

	// handle tx priority: payCoin is denominated in the fee denom (any other
	// denom was already rejected by GetCurrentGasPrice above), so priority is
	// computed directly against the fee-denom gas price. No resolver round
	// trip — with the v30 ErrorDenomResolver a second resolution against a
	// bond denom that differs from the fee denom would reject every tx.
	var priority int64
	if !simulate {
		priority = GetTxPriority(payCoin, int64(gas), feeGasPrice)
	}
	ctx = ctx.WithPriority(priority)

	return next(ctx, tx, simulate)
}

// isBypassMinFeeTx returns true when every message in the tx (recursing into
// authz.MsgExec like MsgFilterDecorator does) is in the bypass-min-fee
// allow-list. An empty tx is not a bypass tx.
func (dfd InnerDeductFeeDecorator) isBypassMinFeeTx(tx sdk.Tx) bool {
	msgs := tx.GetMsgs()
	if len(msgs) == 0 {
		return false
	}

	return dfd.allMsgsBypassMinFee(msgs, 0)
}

// maxBypassRecursionDepth caps authz.MsgExec nesting so a hostile tx cannot
// stack MsgExec wrappers to burn unmetered decode work in CheckTx.
const maxBypassRecursionDepth = 3

func (dfd InnerDeductFeeDecorator) allMsgsBypassMinFee(msgs []sdk.Msg, depth int) bool {
	if depth > maxBypassRecursionDepth {
		return false
	}

	for _, msg := range msgs {
		if exec, ok := msg.(*authz.MsgExec); ok {
			inner, err := exec.GetMessages()
			if err != nil || len(inner) == 0 {
				return false
			}
			if !dfd.allMsgsBypassMinFee(inner, depth+1) {
				return false
			}
			continue
		}

		if !dfd.isBypassMsg(msg) {
			return false
		}
	}

	return true
}

func (dfd InnerDeductFeeDecorator) isBypassMsg(msg sdk.Msg) bool {
	msgType := sdk.MsgTypeURL(msg)
	for _, allowed := range dfd.bypassMinFeeMsgTypes {
		if msgType == allowed {
			return true
		}
	}

	return false
}

// Handle zero fee transactions for x/feepay module.
// CONTRACT: the tx was validated by IsValidFeePayTransaction, which enforces
// exactly one message of type MsgExecuteContract on a registered contract.
func (dfd InnerDeductFeeDecorator) handleZeroFees(ctx sdk.Context, deductFeesFromAcc sdk.AccountI, tx sdk.FeeTx) error {
	msg := tx.GetMsgs()[0]
	cw, ok := msg.(*wasmtypes.MsgExecuteContract)
	if !ok {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "feepay tx message must be a MsgExecuteContract, got %T", msg)
	}

	// Get the fee pay contract
	feepayContract, err := dfd.feepayKeeper.GetContract(ctx, cw.GetContract())
	if err != nil {
		return errorsmod.Wrapf(err, "error getting contract %s", cw.GetContract())
	}

	// Get the fee price in the chain denom
	fmMinGasPriceBondDenom, err := dfd.feemarketKeeper.GetCurrentGasPrice(ctx, dfd.bondDenom)
	if err != nil {
		return errorsmod.Wrapf(err, "error getting feemarket params")
	}
	feePrice := sdk.DecCoin{}
	if fmMinGasPriceBondDenom.Denom == dfd.bondDenom {
		feePrice = fmMinGasPriceBondDenom
	}

	if feePrice == (sdk.DecCoin{}) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "fee price not found for denom %s in feemarket keeper", dfd.bondDenom)
	}

	gas := sdkmath.LegacyNewDec(int64(tx.GetGas()))
	requiredFee := feePrice.Amount.Mul(gas).Ceil().RoundInt()

	// Check if wallet exceeded usage limit on contract
	accBech32 := deductFeesFromAcc.GetAddress().String()
	if dfd.feepayKeeper.HasWalletExceededUsageLimit(ctx, feepayContract, accBech32) {
		return errorsmod.Wrapf(feepaytypes.ErrWalletExceededUsageLimit, "wallet has exceeded usage limit (%d)", feepayContract.WalletLimit)
	}

	// Check if the contract has enough funds to cover the fee
	if !dfd.feepayKeeper.CanContractCoverFee(feepayContract, requiredFee.Uint64()) {
		return errorsmod.Wrapf(feepaytypes.ErrContractNotEnoughFunds, "contract has insufficient funds; expected: %d, got: %d", requiredFee.Uint64(), feepayContract.Balance)
	}

	// Create an array of coins, storing the required fee
	payment := sdk.NewCoins(sdk.NewCoin(feePrice.Denom, requiredFee))

	// Cover the fees of the transaction, send from FeePay Module to FeeCollector Module
	if err := dfd.bankKeeper.SendCoinsFromModuleToModule(ctx, feepaytypes.ModuleName, feemarkettypes.FeeCollectorName, payment); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, "error transferring funds from FeePay to FeeCollector; %s", err)
	}

	// Deduct the fee from the contract balance
	dfd.feepayKeeper.SetContractBalance(ctx, feepayContract, feepayContract.Balance-requiredFee.Uint64())

	// Increment wallet usage
	if err := dfd.feepayKeeper.IncrementContractUses(ctx, feepayContract, accBech32, 1); err != nil {
		return errorsmod.Wrapf(err, "error incrementing contract uses")
	}

	return nil
}

// escrow deducts coins to the escrow.
func (dfd InnerDeductFeeDecorator) escrow(ctx sdk.Context, acc sdk.AccountI, coins sdk.Coins) error {
	targetModuleAcc := feemarkettypes.FeeCollectorName
	err := dfd.bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), targetModuleAcc, coins)
	if err != nil {
		return err
	}

	return nil
}

// CheckTxFee implements the logic for the fee market to check if a Tx has provided sufficient
// fees given the current state of the fee market. Returns an error if insufficient fees.
func CheckTxFee(ctx sdk.Context, gasPrice sdk.DecCoin, feeCoin sdk.Coin, feeGas int64, isAnte bool) (payCoin sdk.Coin, tip sdk.Coin, err error) {
	payCoin = feeCoin

	// Ensure that the provided fees meet the minimum
	if !gasPrice.IsZero() {
		var (
			requiredFee sdk.Coin
			consumedFee sdk.Coin
		)

		// Determine the required fees by multiplying each required minimum gas
		// price by the gas, where fee = ceil(minGasPrice * gas).
		gasConsumed := int64(ctx.GasMeter().GasConsumed())
		gcDec := sdkmath.LegacyNewDec(gasConsumed)
		glDec := sdkmath.LegacyNewDec(feeGas)

		consumedFeeAmount := gasPrice.Amount.Mul(gcDec)
		limitFee := gasPrice.Amount.Mul(glDec)

		consumedFee = sdk.NewCoin(gasPrice.Denom, consumedFeeAmount.Ceil().RoundInt())
		requiredFee = sdk.NewCoin(gasPrice.Denom, limitFee.Ceil().RoundInt())

		if !payCoin.IsGTE(requiredFee) {
			return sdk.Coin{}, sdk.Coin{}, sdkerrors.ErrInsufficientFee.Wrapf(
				"got: %s required: %s, minGasPrice: %s, gas: %d",
				payCoin,
				requiredFee,
				gasPrice,
				gasConsumed,
			)
		}

		if isAnte {
			tip = payCoin.Sub(requiredFee)
			payCoin = requiredFee
		} else {
			tip = payCoin.Sub(consumedFee)
			payCoin = consumedFee
		}
	}

	return payCoin, tip, nil
}

// GetTxPriority returns a naive tx priority based on the amount of gas price provided in a transaction.
//
// The fee amount is divided by the gasLimit to calculate "Effective Gas Price".
// This value is then normalized and scaled into an integer, so it can be used as a priority.
//
//	effectiveGasPrice = feeAmount / gas limit (denominated in fee per gas)
//	normalizedGasPrice = effectiveGasPrice / currentGasPrice (floor is 1.  The minimum effective gas price can ever be is current gas price)
//	scaledGasPrice = normalizedGasPrice * gasPricePrecisionMultiplier (10^6 — decimal places in the normalized gas price to consider when converting to int64).
func GetTxPriority(fee sdk.Coin, gasLimit int64, currentGasPrice sdk.DecCoin) int64 {
	// protections from dividing by 0
	if gasLimit == 0 {
		return 0
	}

	// if the gas price is 0, just use a raw amount
	if currentGasPrice.IsZero() {
		return fee.Amount.Int64()
	}

	effectiveGasPrice := fee.Amount.ToLegacyDec().QuoInt64(gasLimit)
	normalizedGasPrice := effectiveGasPrice.Quo(currentGasPrice.Amount)
	scaledGasPrice := normalizedGasPrice.MulInt64(gasPricePrecisionMultiplier)

	// overflow panic protection
	if scaledGasPrice.GTE(sdkmath.LegacyNewDec(math.MaxInt64)) {
		return math.MaxInt64
	} else if scaledGasPrice.LTE(sdkmath.LegacyOneDec()) {
		return 0
	}

	return scaledGasPrice.TruncateInt64()
}
