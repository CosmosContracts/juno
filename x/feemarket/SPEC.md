# `x/feemarket`

## Contents

- [`x/feemarket`](#xfeemarket)
  - [Contents](#contents)
  - [State](#state)
    - [GasPrice](#gasprice)
    - [LearningRate](#learningrate)
    - [Window](#window)
    - [Index](#index)
  - [Keeper](#keeper)
  - [Messages](#messages)
    - [MsgParams](#msgparams)
  - [Events](#events)
    - [FeePay](#feepay)
    - [TipPay](#tippay)
  - [Parameters](#parameters)
    - [Alpha](#alpha)
    - [Beta](#beta)
    - [Theta](#theta)
    - [Delta](#delta)
    - [MinBaseGasPrice](#minbasegasprice)
    - [MinLearningRate](#minlearningrate)
    - [MaxLearningRate](#maxlearningrate)
    - [MaxBlockUtilization](#maxblockutilization)
    - [Window](#window-1)
    - [FeeDenom](#feedenom)
    - [Enabled](#enabled)
  - [Client](#client)
    - [CLI](#cli)
      - [Query](#query)
        - [params](#params)
        - [state](#state-1)
        - [gas-price](#gas-price)
        - [gas-prices](#gas-prices)
  - [gRPC](#grpc)
    - [Params](#params-1)
    - [State](#state-2)
    - [GasPrice](#gasprice-1)
    - [GasPrices](#gasprices)

## State

The `x/feemarket` module keeps state of the following primary objects:

1. Current base-fee
2. Current learning rate
3. Moving window of block utilization

In addition, the `x/feemarket` module keeps the following indexes to manage the
aforementioned state:

* State: `0x02 |ProtocolBuffer(State)`

### GasPrice

GasPrice is the current gas price. This is denominated in the fee per gas
unit in the base fee denom.

### LearningRate

LearningRate is the current learning rate.

### Window

Window contains a list of the last blocks' utilization values. This is used
to calculate the next base fee. This stores the number of units of gas
consumed per block.

### Index

Index is the index of the current block in the block utilization window.

```protobuf
// State is utilized to track the current state of the fee market. This includes
// the current base fee, learning rate, and block utilization within the
// specified AIMD window.
message State {
  // BaseGasPrice is the current base fee. This is denominated in the fee per gas
  // unit.
  string base_gas_price = 1 [
    (cosmos_proto.scalar) = "cosmos.Dec",
    (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
    (gogoproto.nullable) = false
  ];

  // LearningRate is the current learning rate.
  string learning_rate = 2 [
    (cosmos_proto.scalar) = "cosmos.Dec",
    (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
    (gogoproto.nullable) = false
  ];

  // Window contains a list of the last blocks' utilization values. This is used
  // to calculate the next base fee. This stores the number of units of gas
  // consumed per block.
  repeated uint64 window = 3;

  // Index is the index of the current block in the block utilization window.
  uint64 index = 4;
}
```

## Keeper

The feemarket module provides a keeper interface for accessing the KVStore.

```go
type FeeMarketKeeper interface {
    // Get the current state from the store.
    GetState(ctx sdk.Context) (types.State, error)

    // Set the state in the store.
    SetState(ctx sdk.Context, state types.State) error

    // Get the current params from the store.
    GetParams(ctx sdk.Context) (types.Params, error)

    // Set the params in the store.
    SetParams(ctx sdk.Context, params types.Params) error

	// Get the minimum gas price for a given denom from the store.
    GetMinGasPrice(ctx sdk.Context, denom string) (sdk.DecCoin, error) {

    // Get the current minimum gas prices from the store.
    GetMinGasPrices(ctx sdk.Context) (sdk.DecCoins, error)
}
```

## Messages

### MsgParams

The `feemarket` module params can be updated through `MsgParams`, which can be done using a governance proposal. The signer will always be the `gov` module account address.

```protobuf
message MsgParams {
  option (cosmos.msg.v1.signer) = "authority";

  // Params defines the new parameters for the feemarket module.
  Params params = 1 [ (gogoproto.nullable) = false ];
  // Authority defines the authority that is updating the feemarket module
  // parameters.
  string authority = 2 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
}
```

The message handling can fail if:

* signer is not the gov module account address.

## Events

The feemarket module emits the following events:

### FeePay

```json
{
  "type": "fee_pay",
  "attributes": [
    {
      "key": "fee",
      "value": "{{sdk.Coins being payed}}",
      "index": true
    },
    {
      "key": "fee_payer",
      "value": "{{sdk.AccAddress paying the fees}}",
      "index": true
    }
  ]
}
```

### TipPay

```json
{
  "type": "tip_pay",
  "attributes": [
    {
      "key": "tip",
      "value": "{{sdk.Coins being payed}}",
      "index": true
    },
    {
      "key": "tip_payer",
      "value": "{{sdk.AccAddress paying the tip}}",
      "index": true
    },
    {
      "key": "tip_payee",
      "value": "{{sdk.AccAddress receiving the tip}}",
      "index": true
    }
  ]
}
```

## Parameters

The feemarket module stores it's params in state with the prefix of `0x01`,
which can be updated with governance or the address with authority.

* Params: `0x01 | ProtocolBuffer(Params)`

The feemarket module contains the following parameters:

### Alpha

Alpha is the amount we added to the learning rate
when it is above or below the target +/- threshold.

### Beta

Beta is the amount we multiplicatively decrease the learning rate
when it is within the target +/- threshold.

### Theta

Theta is the threshold for the learning rate. If the learning rate is
above or below the target +/- threshold, we additively increase the
learning rate by Alpha. Otherwise, we multiplicatively decrease the
learning rate by Beta.

### Delta

Delta is the amount we additively increase/decrease the base fee when the
net block utilization difference in the window is above/below the target
utilization.

### MinBaseGasPrice

MinBaseGasPrice determines the initial gas price of the module and the global
minimum for the network. This is denominated in fee per gas unit in the `FeeDenom`.

### MinLearningRate

MinLearningRate is the lower bound for the learning rate.

### MaxLearningRate

MaxLearningRate is the upper bound for the learning rate.


### MaxBlockUtilization

MaxBlockUtilization is the maximum block utilization. Once this has been surpassed,
no more transactions will be added to the current block.

### Window

Window defines the window size for calculating an adaptive learning rate
over a moving window of blocks. The default EIP1559 implementation uses
a window of size 1.

### FeeDenom

FeeDenom is the denom that will be used for all fee payments.

### Enabled

Enabled is a boolean that determines whether the EIP1559 fee market is
enabled. This can be used to add the feemarket module and enable it
through governance at a later time.

```protobuf
// Params contains the required set of parameters for the EIP1559 fee market
// plugin implementation.
message Params {
  // Alpha is the amount we additively increase the learning rate
  // when it is above or below the target +/- threshold.
  //
  // Must be > 0.
  string alpha = 1 [
    (cosmos_proto.scalar) = "cosmos.Dec",
    (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
    (gogoproto.nullable) = false
  ];

  // Beta is the amount we multiplicatively decrease the learning rate
  // when it is within the target +/- threshold.
  //
  // Must be [0, 1].
  string beta = 2 [
    (cosmos_proto.scalar) = "cosmos.Dec",
    (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
    (gogoproto.nullable) = false
  ];

  // Gamma is the threshold for the learning rate. If the learning rate is
  // above or below the target +/- threshold, we additively increase the
  // learning rate by Alpha. Otherwise, we multiplicatively decrease the
  // learning rate by Beta.
  //
  // Must be [0, 0.5].
  string gamma = 3 [
    (cosmos_proto.scalar) = "cosmos.Dec",
    (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
    (gogoproto.nullable) = false
  ];

  // Delta is the amount we additively increase/decrease the gas price when the
  // net block utilization difference in the window is above/below the target
  // utilization.
  string delta = 4 [
    (cosmos_proto.scalar) = "cosmos.Dec",
    (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
    (gogoproto.nullable) = false
  ];

  // MinBaseGasPrice determines the initial gas price of the module and the
  // global minimum for the network.
  string min_base_gas_price = 5 [
    (cosmos_proto.scalar) = "cosmos.Dec",
    (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
    (gogoproto.nullable) = false
  ];

  // MinLearningRate is the lower bound for the learning rate.
  string min_learning_rate = 6 [
    (cosmos_proto.scalar) = "cosmos.Dec",
    (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
    (gogoproto.nullable) = false
  ];

  // MaxLearningRate is the upper bound for the learning rate.
  string max_learning_rate = 7 [
    (cosmos_proto.scalar) = "cosmos.Dec",
    (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
    (gogoproto.nullable) = false
  ];

  // MaxBlockUtilization is the maximum block utilization.
  uint64 max_block_utilization = 8;

  // Window defines the window size for calculating an adaptive learning rate
  // over a moving window of blocks.
  uint64 window = 9;

  // FeeDenom is the denom that will be used for all fee payments.
  string fee_denom = 10;

  // Enabled is a boolean that determines whether the EIP1559 fee market is
  // enabled.
  bool enabled = 11;

  // DistributeFees is a boolean that determines whether the fees are burned or
  // distributed to all stakers.
  bool distribute_fees = 12;
}
```

## Client

### CLI

A user can query and interact with the `feemarket` module using the CLI.

#### Query

The `query` commands allow users to query `feemarket` state.

```shell
feemarketd query feemarket --help
```

##### params

The `params` command allows users to query the on-chain parameters.

```shell
feemarketd query feemarket params [flags]
```

Example:

```shell
feemarketd query feemarket params
```

Example Output:

```yml
alpha: "0.000000000000000000"
beta: "1.000000000000000000"
delta: "0.000000000000000000"
enabled: true
fee_denom: skip
max_block_utilization: "30000000"
max_learning_rate: "0.125000000000000000"
min_base_fee: "1.000000000000000000"
min_learning_rate: "0.125000000000000000"
theta: "0.000000000000000000"
window: "1"
```

##### state

The `state` command allows users to query the current on-chain state.

```shell
feemarketd query feemarket state [flags]
```

Example:

```shell
feemarketd query feemarket state
```

Example Output:

```yml
base_fee: "1.000000000000000000"
index: "0"
learning_rate: "0.125000000000000000"
window:
  - "0"
```

##### gas-price

The `gas-price` command allows users to query the current gas-price for a given denom.

```shell
feemarketd query feemarket gas-price [denom ][flags]
```

Example:

```shell
feemarketd query feemarket gas-price skip
```

Example Output:

```yml
1000000skip
```

##### gas-prices

The `gas-prices` command allows users to query the current gas-price for all supported denoms.

```shell
feemarketd query feemarket gas-prices [flags]
```

Example:

```shell
feemarketd query feemarket gas-prices
```

Example Output:

```yml
1000000stake,100000skip
```

## gRPC

A user can query the `feemarket` module using gRPC endpoints.

### Params

The `Params` endpoint allows users to query the on-chain parameters.

```shell
feemarket.feemarket.v1.Query/Params
```

Example:

```shell
grpcurl -plaintext \
    localhost:9090 \
    feemarket.feemarket.v1.Query/Params
```

Example Output:

```json
{
  "params": {
    "alpha": "0",
    "beta": "1000000000000000000",
    "theta": "0",
    "delta": "0",
    "minBaseFee": "1000000",
    "minLearningRate": "125000000000000000",
    "maxLearningRate": "125000000000000000",
    "maxBlockUtilization": "30000000",
    "window": "1",
    "feeDenom": "skip",
    "enabled": true
  }
}
```

### State

The `State` endpoint allows users to query the current on-chain state.

```shell
feemarket.feemarket.v1.Query/State
```

Example:

```shell
grpcurl -plaintext \
    localhost:9090 \
    feemarket.feemarket.v1.Query/State
```

Example Output:

```json
{
  "state": {
    "baseGasPrice": "1000000",
    "learningRate": "125000000000000000",
    "window": [
      "0"
    ]
  }
}
```

### GasPrice

The `GasPrice` endpoint allows users to query the current on-chain gas price for a given denom.

```shell
feemarket.feemarket.v1.Query/GasPrice
```

Example:

```shell
grpcurl -plaintext \
    -d '{"denom": "skip"}' \
    localhost:9090 \
    feemarket.feemarket.v1.Query/GasPrice/
```

Example Output:

```json
{
  "price": {
      "denom": "skip",
      "amount": "1000000"
  }
}
```

### GasPrices

The `GasPrices` endpoint allows users to query the current on-chain gas prices for all denoms.

```shell
feemarket.feemarket.v1.Query/GasPrices
```

Example:

```shell
grpcurl -plaintext \
    localhost:9090 \
    feemarket.feemarket.v1.Query/GasPrices
```

Example Output:

```json
{
  "prices": [
    {
      "denom": "skip",
      "amount": "1000000"
    }
  ]
}
```