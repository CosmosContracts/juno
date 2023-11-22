<!--
order: 2
-->

# State

## State Objects

The `x/feepay` module keeps the following objects in the state: FeePayContract and FeePayWalletUsage. These objects are used to store the state of a contract and the number of times a wallet has interacted with a contract.

```go
// This defines the address, balance, and wallet limit
// of a fee pay contract.
message FeePayContract {  
  // The address of the contract.
  string contract_address = 1;
  // The ledger balance of the contract.
  uint64 balance = 2;
  // The number of times a wallet may interact with the contract.
  uint64 wallet_limit = 3;
}
```

```go
// This object is used to store the number of times a wallet has
// interacted with a contract.
message FeePayWalletUsage {
  // The contract address.
  string contract_address = 1;
  // The wallet address.
  string wallet_address = 2;
  // The number of uses corresponding to a wallet.
  uint64 uses = 3;
}
```

## Genesis & Params

The `x/feepay` module's `GenesisState` defines the state necessary for initializing the chain from a previously exported height. It contains the module parameters and the fee pay contracts. As of now, it does not contain the wallet usage. The params are used to enable or disable the module. This value can be modified with a governance proposal.

```go
// GenesisState defines the module's genesis state.
message GenesisState {
  // params are the feepay module parameters
  Params params = 1 [ (gogoproto.nullable) = false ];

  // fee_pay_contracts are the feepay module contracts
  repeated FeePayContract fee_pay_contracts = 2 [ (gogoproto.nullable) = false ];
}

// Params defines the feepay module params
message Params {
  // enable_feepay defines a parameter to enable the feepay module
  bool enable_feepay = 1;
}
```

## State Transitions

The following state transitions are possible:

- Registering a contract creates a FeePayContract object in the state.
- Unregistering a contract removes the FeePayContract object from the state.
- Funding a contract updates the balance of the FeePayContract object in the state.
- Updating the wallet limit of a contract updates the FeePayContract object in the state.
- Interacting with a contract updates the FeePayWalletUsage object in the state and deducts the balance of the FeePayContract object in the state.