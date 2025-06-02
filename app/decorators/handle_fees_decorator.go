package decorators

import (
	"bytes"
	"fmt"
	"math"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	feegrantKeeper "cosmossdk.io/x/feegrant/keeper"

	feemarketkeeper "github.com/CosmosContracts/juno/v30/x/feemarket/keeper"
	feemarkettypes "github.com/CosmosContracts/juno/v30/x/feemarket/types"
	feepayhelpers "github.com/CosmosContracts/juno/v30/x/feepay/helpers"
	feepaykeeper "github.com/CosmosContracts/juno/v30/x/feepay/keeper"
	feepaytypes "github.com/CosmosContracts/juno/v30/x/feepay/types"
)

const (
	// gasPricePrecision is the amount of digit precision to scale the gas prices to.
	gasPricePrecision = 6
)

type DeductFeeDecorator struct {
	feemarketkeeper feemarketkeeper.Keeper

	innerDecorator    InnerDeductFeeDecorator
	fallbackDecorator sdk.AnteDecorator
}

func NewDeductFeeDecorator(fpk feepaykeeper.Keeper, fmk feemarketkeeper.Keeper, ak authkeeper.AccountKeeper, bk bankkeeper.Keeper, fgk feegrantKeeper.Keeper, bondDenom string, fallbackDecorator sdk.AnteDecorator) DeductFeeDecorator {
	return DeductFeeDecorator{
		feemarketkeeper: fmk,
		innerDecorator: newInnerDeductFeeDecorator(
			fpk, fmk, ak, bk, fgk, bondDenom,
		),
		fallbackDecorator: fallbackDecorator,
	}
}

// DeductFeeDecorator deducts fees from the first signer of the tx
// If the first signer does not have the funds to pay for the fees, return with InsufficientFunds error
// Call next AnteHandler if fees successfully deducted
// CONTRACT: Tx must implement FeeTx interface to use DeductFeeDecorator
//
// Additionally, the Deduct Fee ante is a fork of the SDK's DeductFeeDecorator. This decorator looks for single
// message transactions with no provided fee. If they correspond to a registered FeePay Contract, the FeePay
// module will cover the cost of the fee (if the balance permits).
type InnerDeductFeeDecorator struct {
	feepayKeeper    feepaykeeper.Keeper
	feemarketKeeper feemarketkeeper.Keeper
	accountKeeper   authkeeper.AccountKeeper
	bankKeeper      bankkeeper.Keeper
	feegrantKeeper  feegrantKeeper.Keeper
	bondDenom       string
}

func newInnerDeductFeeDecorator(fpk feepaykeeper.Keeper, fmk feemarketkeeper.Keeper, ak authkeeper.AccountKeeper, bk bankkeeper.Keeper, fgk feegrantKeeper.Keeper, bondDenom string) InnerDeductFeeDecorator {
	return InnerDeductFeeDecorator{
		feepayKeeper:    fpk,
		feemarketKeeper: fmk,
		accountKeeper:   ak,
		bankKeeper:      bk,
		feegrantKeeper:  fgk,
		bondDenom:       bondDenom,
	}
}

// AnteHandle calls the feemarket internal antehandler if the keeper is enabled.  If disabled, the fallback
// fee antehandler is fallen back to.
func (dfd DeductFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	params, err := dfd.feemarketkeeper.GetParams(ctx)
	if err != nil {
		return ctx, err
	}
	if params.Enabled {
		return dfd.innerDecorator.anteHandle(ctx, tx, simulate, next)
	}

	// only use fallback if not nil
	if dfd.fallbackDecorator != nil {
		return dfd.fallbackDecorator.AnteHandle(ctx, tx, simulate, next)
	}

	return next(ctx, tx, simulate)
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
			return ctx, errorsmod.Wrapf(feemarkettypes.ErrNoFeeCoins, "got length %d", len(feeCoins))
		}
	}

	if len(feeCoins) > 1 {
		return ctx, errorsmod.Wrapf(feemarkettypes.ErrTooManyFeeCoins, "got length %d", len(feeCoins))
	}

	payCoin := sdk.NewCoin(dfd.bondDenom, sdkmath.ZeroInt())
	if !simulate {
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

	if !simulate {
		_, _, err := CheckTxFee(ctx, feeGasPrice, payCoin, int64(gas), true)
		if err != nil {
			return ctx, errorsmod.Wrapf(err, "error checking fee")
		}
	}

	// handle the entire tx fee process
	err = dfd.HandleFees(ctx, feeTx, payCoin, isValidFeepayTx)
	if err != nil {
		return ctx, errorsmod.Wrapf(err, "error escrowing funds")
	}

	// handle tx priority
	var priority int64 = 0
	bondDenomGasPrice, err := dfd.feemarketKeeper.GetCurrentGasPrice(ctx, dfd.bondDenom)
	priorityFee, err := dfd.resolveTxPriorityCoins(ctx, payCoin, dfd.bondDenom)
	if err != nil {
		return ctx, errorsmod.Wrapf(err, "error resolving fee priority")
	}
	if !simulate {
		priority = GetTxPriority(priorityFee, int64(gas), bondDenomGasPrice)
	}
	ctx = ctx.WithPriority(priority)

	return next(ctx, tx, simulate)
}

func (dfd InnerDeductFeeDecorator) HandleFees(ctx sdk.Context, feeTx sdk.FeeTx, fee sdk.Coin, isValidFeepayTx bool) error {
	if addr := dfd.accountKeeper.GetModuleAddress(authtypes.FeeCollectorName); addr == nil {
		return fmt.Errorf("fee collector module account (%s) has not been set", authtypes.FeeCollectorName)
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

// Handle zero fee transactions for fee prepay module
func (dfd InnerDeductFeeDecorator) handleZeroFees(ctx sdk.Context, deductFeesFromAcc sdk.AccountI, tx sdk.FeeTx) error {
	msg := tx.GetMsgs()[0]
	cw := msg.(*wasmtypes.MsgExecuteContract)

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

// resolveTxPriorityCoins converts the coins to the proper denom used for tx prioritization calculation.
func (dfd InnerDeductFeeDecorator) resolveTxPriorityCoins(ctx sdk.Context, fee sdk.Coin, baseDenom string) (sdk.Coin, error) {
	if fee.Denom == baseDenom {
		return fee, nil
	}

	feeDec := sdk.NewDecCoinFromCoin(fee)
	convertedDec, err := dfd.feemarketKeeper.ResolveToDenom(ctx, feeDec, baseDenom)
	if err != nil {
		return sdk.Coin{}, err
	}

	// truncate down
	return sdk.NewCoin(baseDenom, convertedDec.Amount.TruncateInt()), nil
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
//	scaledGasPrice = normalizedGasPrice * 10 ^ gasPricePrecision (amount of decimal places in the normalized gas price to consider when converting to int64).
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
	scaledGasPrice := normalizedGasPrice.MulInt64(int64(math.Pow10(gasPricePrecision)))

	// overflow panic protection
	if scaledGasPrice.GTE(sdkmath.LegacyNewDec(math.MaxInt64)) {
		return math.MaxInt64
	} else if scaledGasPrice.LTE(sdkmath.LegacyOneDec()) {
		return 0
	}

	return scaledGasPrice.TruncateInt64()
}
