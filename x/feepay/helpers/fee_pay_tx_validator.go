package helpers

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	feepaykeeper "github.com/CosmosContracts/juno/v18/x/feepay/keeper"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

// Check if a transaction should be processed as a FeePay transaction. Ensure that
// the Fee Pay module is enabled before calling this function. A valid FeePay transaction
// has no fee, and only 1 message for executing a contract.
//
// TODO: Future allow for multiple msgs.
func IsValidFeePayTransaction(ctx sdk.Context, feePayKeeper feepaykeeper.Keeper, tx sdk.Tx, fee sdk.Coins) bool {

	// Defaults to false
	isValid := false

	// Check if the fee pay module is enabled
	isEnabled := feePayKeeper.GetParams(ctx).EnableFeepay

	// Check if fee is zero, and tx has only 1 message for executing a contract
	if isEnabled && fee.IsZero() && len(tx.GetMsgs()) == 1 {
		_, isValid = (tx.GetMsgs()[0]).(*wasmtypes.MsgExecuteContract)
	}

	// Return if the tx is valid
	return isValid
}
