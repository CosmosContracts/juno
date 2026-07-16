package ante

import (
	"encoding/json"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	feemarkettypes "github.com/CosmosContracts/juno/v30/x/feemarket/types"
	"github.com/CosmosContracts/juno/v30/x/feeshare/keeper"
	"github.com/CosmosContracts/juno/v30/x/feeshare/types"
)

// FeeSharePayoutDecorator Run his after we already deduct the fee from the account with
// the ante.NewDeductFeeDecorator() decorator. We pull funds from the FeeCollector ModuleAccount
type FeeSharePayoutDecorator struct {
	bankKeeper     bankkeeper.Keeper
	feesharekeeper keeper.Keeper
}

func NewFeeSharePayoutDecorator(bk bankkeeper.Keeper, fs keeper.Keeper) FeeSharePayoutDecorator {
	return FeeSharePayoutDecorator{
		bankKeeper:     bk,
		feesharekeeper: fs,
	}
}

func (fsd FeeSharePayoutDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	// In simulate mode the v30 DeductFeeDecorator does not escrow into
	// feemarket-fee-collector (payCoin is zero in simulate, so the
	// `else if !fee.IsZero()` branch in HandleFees is skipped). Running the
	// payout here would then read 0 balance and fail simulate, breaking
	// `--gas auto` for every feeshare-registered contract execute. Skip the
	// payout in simulate; --gas-adjustment provides headroom for the small
	// amount of gas the bank send would consume.
	if simulate {
		return next(ctx, tx, simulate)
	}

	err = fsd.FeeSharePayout(ctx, fsd.bankKeeper, feeTx.GetFee(), fsd.feesharekeeper, tx.GetMsgs())
	if err != nil {
		return ctx, errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, "%s", err.Error())
	}

	return next(ctx, tx, simulate)
}

// CalculateFeeSharePool computes the total developer pool per denom:
// RoundInt(govPercent * feeAmount). The pool is computed ONCE for the whole
// tx — never per recipient — so the aggregate paid out can never exceed
// govPercent of the fee.
// tested in ante_test.go
func CalculateFeeSharePool(fees sdk.Coins, govPercent sdkmath.LegacyDec) sdk.Coins {
	var pool sdk.Coins
	for _, c := range fees.Sort() {
		poolAmount := govPercent.MulInt(c.Amount).RoundInt()
		if !poolAmount.IsZero() {
			pool = pool.Add(sdk.NewCoin(c.Denom, poolAmount))
		}
	}
	return pool
}

// SplitFeeSharePool splits the developer pool between numPairs recipients with
// a running remainder: everyone gets pool/numPairs (truncated) and the LAST
// recipient additionally receives the leftover. The aggregate always equals
// the pool exactly — the old per-recipient RoundInt(pool/numPairs) could sum
// to MORE than the pool and overdraw the escrow, reverting the tx.
// tested in ante_test.go
func SplitFeeSharePool(pool sdk.Coins, numPairs int) []sdk.Coins {
	splits := make([]sdk.Coins, numPairs)
	if numPairs == 0 {
		return splits
	}

	for _, c := range pool {
		share := c.Amount.QuoRaw(int64(numPairs))
		remainder := c.Amount.Sub(share.MulRaw(int64(numPairs)))

		for i := range numPairs {
			amount := share
			if i == numPairs-1 {
				amount = amount.Add(remainder)
			}
			if !amount.IsZero() {
				splits[i] = splits[i].Add(sdk.NewCoin(c.Denom, amount))
			}
		}
	}

	return splits
}

type FeeSharePayoutEventOutput struct {
	WithdrawAddress sdk.AccAddress `json:"withdraw_address"`
	FeesPaid        sdk.Coins      `json:"fees_paid"`
}

// Loop through all messages and add the withdraw address to the list of addresses to pay
// if the contract opted-in to fee sharing
func addNewFeeSharePayoutsForMsgs(ctx sdk.Context, fsk keeper.Keeper, toPay *[]sdk.AccAddress, msgs []sdk.Msg) error {
	// Check if an authz message, loop through all inner messages, and recursively call this function
	for _, msg := range msgs {
		if authzMsg, ok := msg.(*authz.MsgExec); ok {
			innerMsgs, err := authzMsg.GetMessages()
			if err != nil {
				return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "cannot unmarshal authz exec msgs")
			}

			// Recursively call this function with the inner messages
			err = addNewFeeSharePayoutsForMsgs(ctx, fsk, toPay, innerMsgs)
			if err != nil {
				return err
			}
		}

		// If an execute contract message, check if the contract opted-in to fee sharing,
		// and if so, add the withdraw address to the list of addresses to pay
		if execContractMsg, ok := msg.(*wasmtypes.MsgExecuteContract); ok {
			contractAddr, err := sdk.AccAddressFromBech32(execContractMsg.Contract)
			if err != nil {
				return err
			}

			shareData, _ := fsk.GetFeeShare(ctx, contractAddr)

			withdrawAddr := shareData.GetWithdrawerAddr()
			if withdrawAddr != nil && !withdrawAddr.Empty() {
				*toPay = append(*toPay, withdrawAddr)
			}
		}
	}

	return nil
}

// FeeSharePayout takes the total fees and redistributes 50% (or param set) to the contract developers
// provided they opted-in to payments.
func (FeeSharePayoutDecorator) FeeSharePayout(ctx sdk.Context, bankKeeper bankkeeper.Keeper, totalFees sdk.Coins, fsk keeper.Keeper, msgs []sdk.Msg) error {
	params := fsk.GetParams(ctx)
	if !params.EnableFeeShare {
		return nil
	}

	// Get valid withdraw addresses from contracts
	toPay := make([]sdk.AccAddress, 0)

	// Add fee share payouts for each msg
	err := addNewFeeSharePayoutsForMsgs(ctx, fsk, &toPay, msgs)
	if err != nil {
		return err
	}

	// Do nothing if no one needs payment
	if len(toPay) == 0 {
		return nil
	}

	// Get only allowed governance fees to be paid (helps for taxes)
	var fees sdk.Coins
	if len(params.AllowedDenoms) == 0 {
		// If empty, we allow all denoms to be used as payment
		fees = totalFees
	} else {
		for _, fee := range totalFees.Sort() {
			for _, allowed := range params.AllowedDenoms {
				if fee.Denom == allowed {
					fees = fees.Add(fee)
				}
			}
		}
	}

	numPairs := len(toPay)

	feesPaidOutput := make([]FeeSharePayoutEventOutput, numPairs)
	if numPairs > 0 {
		govPercent := params.DeveloperShares

		// Compute the developer pool ONCE (govPercent of the fee), then clamp
		// it to what is actually escrowed so the payout can never overdraw
		// the module account and revert the tx.
		pool := CalculateFeeSharePool(fees, govPercent)
		collectorAddr := authtypes.NewModuleAddress(feemarkettypes.FeeCollectorName)
		for i, c := range pool {
			escrowed := bankKeeper.GetBalance(ctx, collectorAddr, c.Denom).Amount
			if c.Amount.GT(escrowed) {
				pool[i].Amount = escrowed
			}
		}

		splits := SplitFeeSharePool(pool, numPairs)

		// pay fees between all withdraw addresses (last recipient absorbs the
		// integer-division remainder). Source from the feemarket fee
		// collector — the v30 DeductFeeDecorator escrows fees to
		// feemarkettypes.FeeCollectorName ("feemarket-fee-collector"),
		// not authtypes.FeeCollectorName ("fee_collector"). The post-handler
		// then drains feemarket-fee-collector after the tx runs. Reading
		// from auth.FeeCollector here returned 0 balance and surfaced as
		// "spendable balance 0ujuno is smaller than 25000ujuno: insufficient
		// funds: feeshare payment error" the moment any feeshare-registered
		// contract was executed.
		for i, withdrawAddr := range toPay {
			err := bankKeeper.SendCoinsFromModuleToAccount(ctx, feemarkettypes.FeeCollectorName, withdrawAddr, splits[i])
			feesPaidOutput[i] = FeeSharePayoutEventOutput{
				WithdrawAddress: withdrawAddr,
				FeesPaid:        splits[i],
			}

			if err != nil {
				return errorsmod.Wrapf(types.ErrFeeSharePayment, "failed to pay fees to contract developer: %s", err.Error())
			}
		}
	}

	bz, err := json.Marshal(feesPaidOutput)
	if err != nil {
		return errorsmod.Wrapf(types.ErrFeeSharePayment, "failed to marshal feesPaidOutput: %s", err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePayoutFeeShare,
			sdk.NewAttribute(types.AttributeWithdrawPayouts, string(bz))),
	)

	return nil
}
