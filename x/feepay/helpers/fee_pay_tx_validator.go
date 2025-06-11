package helpers

import (
	"context"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	feepaykeeper "github.com/CosmosContracts/juno/v30/x/feepay/keeper"
)

// Check if a transaction should be processed as a FeePay transaction.
// A valid FeePay transaction has no fee and contains only messages which
// execute registered CW contracts.
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

	// Check if transaction has at least one message
	msgs := feeTx.GetMsgs()
	if len(msgs) == 0 {
		return false
	}

	// Check that all messages are CW contract executions on registered contracts
	for _, msg := range msgs {
		// Check if the message is a CW contract execution
		cw, ok := msg.(*wasmtypes.MsgExecuteContract)
		if !ok {
			return false
		}

		// Check if the contract is registered
		if _, err := feePayKeeper.GetContract(ctx, cw.Contract); err != nil {
			return false
		}
	}

	return true
}
