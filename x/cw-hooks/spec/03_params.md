# Parameters

The cw-hooks module contains the following parameters:

| Key                        | Type        | Default Value    |
| :------------------------- | :---------- | :--------------- |
| `ContractGasLimit`         | uint64      | `250_000`        |

## Contract Gas Limit

The `ContractGasLimit` parameter is the maximum amount of gas that can be used by a contract in a single event. This is to prevent malicious contracts from spamming the network since all executes are feeless. If you need to perform more than 250,000 Gas execution, you can submit a proposal to increase this for the chain.
