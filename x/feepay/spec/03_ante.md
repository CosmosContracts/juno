<!--
order: 3
-->

# Ante

The `x/feepay` module implements two handlers to perform all of the necessary logic on incoming transactions.

## FeeRouteDecorator Logic

The FeeRouteDecorator is responsible for determining if a transaction is to be processed as a FeePay transaction and correctly routes it to additional decorators for further processing. Below are the steps taken by the FeeRouteDecorator to process a transaction:

1. Flag incoming transaction as a FeePay transaction or not a FeePay transaction (Requirements: 0 provided gas, 0 provided fee, contains only 1 message, only message is a MsgExecuteContract message, & the contract is registered with FeePay)
2. If a FeePay transaction: 
   1. Route to FeePayDecorator (If an error occurs: Handle transaction normally with the SDK's DeductFeeDecorator logic & proceed if no additional errors occur)
   2. Route to GlobalFeeDecorator
3. If not a FeePay transaction: 
   1. Route to GlobalFeeDecorator
   2. Route to FeePayDecorator

The purpose of routing a FeePay transaction to the FeePayDecorator first is to avoid the GlobalFeeDecorator from erroring out due to no provided fees. Additionally, if the FeePayDecorator logic fails, the transaction will attempt to be processed by the SDK's DeductFeeDecorator logic and then proceed to the GlobalFeeDecorator if no errors occur.

## FeePayDecorator Logic

> Note: The FeePayDecorator is an extension of the SDK's DeductFeeDecorator.

The FeePayDecorator is responsible for handling normal fee deductions (on normal transactions) and all FeePay transaction logic (on FeePay transactions). Below are the steps taken by the FeePayDecorator to process a transaction:

1. If not a FeePay transaction: 
   1. Deduct fees from the transaction normally, just like the default SDK decorator
2. If a FeePay transaction:
   1. Determine the required fee to cover the contract execution gas cost
   2. Ensure wallet has not exceeded limit
   3. Ensure contract has enough funds to cover fee
   4. Transfer funds to the FeeCollector module from the contract's funds
   5. Update contract funds in state
   6. Increment wallet usage in state

If any of the FeePay transaction steps fail, the transaction will attempt to be processed normally by the SDK's DeductFeeDecorator logic. If the fallback attempt succeeds, the transaction will pass. If the fallback attempt fails, the transaction will fail and the client will be notified of any and all errors.