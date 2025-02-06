package ante

import (
	"context"
	"errors"

	tmstrings "github.com/cometbft/cometbft/libs/strings"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	globalfeekeeper "github.com/CosmosContracts/juno/v27/x/globalfee/keeper"
)

// FeeWithBypassDecorator checks if the transaction's fee is at least as large
// as the local validator's minimum gasFee (defined in validator config) and global fee, and the fee denom should be in the global fees' denoms.
//
// If fee is too low, decorator returns error and tx is rejected from mempool.
// Note this only applies when ctx.CheckTx = true. If fee is high enough or not
// CheckTx, then call next AnteHandler.
//
// CONTRACT: Tx must implement FeeTx to use FeeDecorator
// If the tx msg type is one of the bypass msg types, the tx is valid even if the min fee is lower than normally required.
// If the bypass tx still carries fees, the fee denom should be the same as global fee required.

var _ sdk.AnteDecorator = FeeDecorator{}

type FeeDecorator struct {
	BypassMinFeeMsgTypes            []string
	GlobalFeeKeeper                 globalfeekeeper.Keeper
	StakingKeeper                   stakingkeeper.Keeper
	MaxTotalBypassMinFeeMsgGasUsage uint64
	IsFeePayTx                      *bool
}

func NewFeeDecorator(bypassMsgTypes []string, gfk globalfeekeeper.Keeper, sk stakingkeeper.Keeper, maxTotalBypassMinFeeMsgGasUsage uint64, isFeePayTx *bool) FeeDecorator {
	return FeeDecorator{
		BypassMinFeeMsgTypes:            bypassMsgTypes,
		GlobalFeeKeeper:                 gfk,
		StakingKeeper:                   sk,
		MaxTotalBypassMinFeeMsgGasUsage: maxTotalBypassMinFeeMsgGasUsage,
		IsFeePayTx:                      isFeePayTx,
	}
}

// AnteHandle implements the AnteDecorator interface
func (mfd FeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must implement the sdk.FeeTx interface")
	}

	// Call next handler if the execution mode is CheckTx, simulation, or if the tx is a fee pay tx
	if !ctx.IsCheckTx() || simulate || *mfd.IsFeePayTx {
		return next(ctx, tx, simulate)
	}

	// Sort fee tx's coins, zero coins in feeCoins are already removed
	feeCoins := feeTx.GetFee().Sort()
	gas := feeTx.GetGas()
	msgs := feeTx.GetMsgs()

	// Get required Global Fee
	requiredGlobalFees, err := mfd.GetGlobalFee(ctx, feeTx)
	if err != nil {
		return ctx, err
	}

	// Get local minimum-gas-prices
	localFees := GetMinGasPrice(ctx, int64(feeTx.GetGas()))

	// CombinedFeeRequirement should never be empty since
	// global fee is set to its default value, i.e. 0uatom, if empty
	combinedFeeRequirement := CombinedFeeRequirement(requiredGlobalFees, localFees)
	if len(combinedFeeRequirement) == 0 {
		return ctx, errorsmod.Wrapf(sdkerrors.ErrNotFound, "required fees are not setup.")
	}

	nonZeroCoinFeesReq, zeroCoinFeesDenomReq := getNonZeroFees(combinedFeeRequirement)

	// feeCoinsNonZeroDenom contains non-zero denominations from the combinedFeeRequirement
	//
	// feeCoinsNoZeroDenom is used to check if the fees meets the requirement imposed by nonZeroCoinFeesReq
	// when feeCoins does not contain zero coins' denoms in combinedFeeRequirement
	feeCoinsNonZeroDenom, feeCoinsZeroDenom := splitCoinsByDenoms(feeCoins, zeroCoinFeesDenomReq)

	// Check that the fees are in expected denominations.
	// if feeCoinsNoZeroDenom=[], DenomsSubsetOf returns true
	// if feeCoinsNoZeroDenom is not empty, but nonZeroCoinFeesReq empty, return false
	if !feeCoinsNonZeroDenom.DenomsSubsetOf(nonZeroCoinFeesReq) {
		return ctx, errorsmod.Wrapf(sdkerrors.ErrInsufficientFee, "this fee denom is not accepted; got %s, one is required: %s", feeCoins, PrettyPrint(combinedFeeRequirement))
	}

	// Accept zero fee transactions only if both of the following statements are true:
	//
	// 	- the tx contains only message types that can bypass the minimum fee,
	//	see BypassMinFeeMsgTypes;
	//	- the total gas limit per message does not exceed MaxTotalBypassMinFeeMsgGasUsage,
	//	i.e., totalGas <=  MaxTotalBypassMinFeeMsgGasUsage
	//
	// Otherwise, minimum fees and global fees are checked to prevent spam.
	doesNotExceedMaxGasUsage := gas <= mfd.MaxTotalBypassMinFeeMsgGasUsage
	allowedToBypassMinFee := mfd.ContainsOnlyBypassMinFeeMsgs(msgs) && doesNotExceedMaxGasUsage

	// Either the transaction contains at least one message of a type
	// that cannot bypass the minimum fee or the total gas limit exceeds
	// the imposed threshold. As a result, besides check the fees are in
	// expected denominations, check the amounts are greater or equal than
	// the expected amounts.

	// only check feeCoinsNoZeroDenom has coins IsAnyGTE than nonZeroCoinFeesReq
	// when feeCoins does not contain denoms of zero denoms in combinedFeeRequirement
	if !allowedToBypassMinFee && len(feeCoinsZeroDenom) == 0 {
		// special case: when feeCoins=[] and there is zero coin in fee requirement
		if len(feeCoins) == 0 && len(zeroCoinFeesDenomReq) != 0 {
			return next(ctx, tx, simulate)
		}

		// Check that the amounts of the fees are greater or equal than
		// the expected amounts, i.e., at least one feeCoin amount must
		// be greater or equal to one of the combined required fees.

		// if feeCoinsNoZeroDenom=[], return false
		// if nonZeroCoinFeesReq=[], return false (this situation should not happen
		// because when nonZeroCoinFeesReq empty, and DenomsSubsetOf check passed,
		// the tx should already passed before)
		if !feeCoinsNonZeroDenom.IsAnyGTE(nonZeroCoinFeesReq) {
			if len(feeCoins) == 0 {
				return ctx, errorsmod.Wrapf(sdkerrors.ErrInsufficientFee, "no fees were specified; one fee must be provided %s", PrettyPrint(combinedFeeRequirement))
			}

			return ctx, errorsmod.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; only got: %s. one is required: %s. ", feeCoins, PrettyPrint(combinedFeeRequirement))
		}
	}

	return next(ctx, tx, simulate)
}

// GetGlobalFee returns the global fees for a given fee tx's gas
// (might also return 0denom if globalMinGasPrice is 0)
// sorted in ascending order.
// Note that ParamStoreKeyMinGasPrices type requires coins sorted.
func (mfd FeeDecorator) GetGlobalFee(ctx context.Context, feeTx sdk.FeeTx) (sdk.Coins, error) {
	var (
		globalMinGasPrices sdk.DecCoins
		err                error
	)

	globalMinGasPrices = mfd.GlobalFeeKeeper.GetParams(ctx).MinimumGasPrices

	// global fee is empty set, set global fee to 0uatom
	if len(globalMinGasPrices) == 0 {
		globalMinGasPrices, err = mfd.DefaultZeroGlobalFee(ctx)
		if err != nil {
			return sdk.Coins{}, err
		}
	}
	requiredGlobalFees := make(sdk.Coins, len(globalMinGasPrices))
	// Determine the required fees by multiplying each required minimum gas
	// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
	glDec := sdkmath.LegacyNewDec(int64(feeTx.GetGas()))
	for i, gp := range globalMinGasPrices {
		fee := gp.Amount.Mul(glDec)
		requiredGlobalFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
	}

	return requiredGlobalFees.Sort(), nil
}

func (mfd FeeDecorator) DefaultZeroGlobalFee(ctx context.Context) ([]sdk.DecCoin, error) {
	bondDenom, err := mfd.getBondDenom(ctx)
	if bondDenom == "" {
		return nil, errors.New("empty staking bond denomination")
	}
	if err != nil {
		return nil, err
	}

	return []sdk.DecCoin{sdk.NewDecCoinFromDec(bondDenom, sdkmath.LegacyNewDec(0))}, nil
}

func (mfd FeeDecorator) getBondDenom(ctx context.Context) (string, error) {
	denom, err := mfd.StakingKeeper.BondDenom(ctx)
	if err != nil {
		return "", err
	}
	return denom, nil
}

// ContainsOnlyBypassMinFeeMsgs returns true if all the given msgs type are listed
// in the BypassMinFeeMsgTypes of the FeeDecorator.
func (mfd FeeDecorator) ContainsOnlyBypassMinFeeMsgs(msgs []sdk.Msg) bool {
	for _, msg := range msgs {
		if tmstrings.StringInSlice(sdk.MsgTypeURL(msg), mfd.BypassMinFeeMsgTypes) {
			continue
		}
		return false
	}

	return true
}

// GetMinGasPrice returns the validator's minimum gas prices
// fees given a gas limit
func GetMinGasPrice(ctx context.Context, gasLimit int64) sdk.Coins {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	minGasPrices := sdkCtx.MinGasPrices()
	// special case: if minGasPrices=[], requiredFees=[]
	if minGasPrices.IsZero() {
		return sdk.Coins{}
	}

	requiredFees := make(sdk.Coins, len(minGasPrices))
	// Determine the required fees by multiplying each required minimum gas
	// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
	glDec := sdkmath.LegacyNewDec(gasLimit)
	for i, gp := range minGasPrices {
		fee := gp.Amount.Mul(glDec)
		requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
	}

	return requiredFees.Sort()
}
