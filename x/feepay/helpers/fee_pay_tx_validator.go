package helpers

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	feepaykeeper "github.com/CosmosContracts/juno/v20/x/feepay/keeper"
)

// Check if a transaction should be processed as a FeePay transaction.
// A valid FeePay transaction has no fee and only 1 message which
// executes a CW contract.
//
// TODO: Future allow for multiple msgs.
func IsValidFeePayTransaction(ctx sdk.Context, feePayKeeper feepaykeeper.Keeper, feeTx sdk.FeeTx) bool {
	// Defaults to false
	isValid := false

	// Check if the fee pay module is enabled
	isEnabled := feePayKeeper.GetParams(ctx).EnableFeepay

	// Check if fee is zero, and tx has only 1 message for executing a contract
	if isEnabled && feeTx.GetFee().IsZero() && len(feeTx.GetMsgs()) == 1 {
		// Check if the message is a CW contract execution
		if cw, ok := (feeTx.GetMsgs()[0]).(*wasmtypes.MsgExecuteContract); ok {
			// Check if the contract is registered
			if _, err := feePayKeeper.GetContract(ctx, cw.Contract); err == nil {
				isValid = true
			}
		}
	}

	// Return if the tx is valid
	return isValid
}
