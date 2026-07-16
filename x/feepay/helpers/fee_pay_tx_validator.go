package helpers

import (
	"context"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	feepaykeeper "github.com/CosmosContracts/juno/v30/x/feepay/keeper"
)

// IsValidFeePayTransaction checks if a transaction should be processed as a FeePay transaction.
// A valid FeePay transaction has no fee attached and contains EXACTLY ONE message,
// which is executing a feepay-registered contract.
//
// The single-message restriction matches the per-tx fee model: the ante handler
// charges (and rate-limits) exactly one contract per tx. Allowing multi-message
// txs would charge only the first message's contract while every message rides
// for free — bypassing both the per-wallet usage limit and the other contracts'
// balances.
func IsValidFeePayTransaction(ctx context.Context, feePayKeeper feepaykeeper.Keeper, feeTx sdk.FeeTx) bool {
	// Check if the fee pay module is enabled
	isEnabled := feePayKeeper.GetParams(ctx).EnableFeepay
	if !isEnabled {
		return false
	}

	// Check if fee is zero
	if !feeTx.GetFee().IsZero() {
		return false
	}

	// A feepay tx must contain exactly one message
	msgs := feeTx.GetMsgs()
	if len(msgs) != 1 {
		return false
	}

	// Check that the message is a CW contract execution
	cw, ok := msgs[0].(*wasmtypes.MsgExecuteContract)
	if !ok {
		return false
	}

	// Check if the contract is registered
	if _, err := feePayKeeper.GetContract(ctx, cw.Contract); err != nil {
		return false
	}

	return true
}
