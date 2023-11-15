<!--
order: 1
-->

# Concepts

## FeePay

The FeePay module provides functionality for Smart Contract developers to cover the execution fees of transactions interacting with their contract. This aims to improve the user experience and help onboard wallets with little to no available funds. Developers can setup their contract with FeePay by first registering it and then funding it with Juno. Clients can then interact with the contract by explicitly specifying 0 fees.

## Registering a Contract

Register a contract with FeePay by executing the following transaction:

```bash
junod tx feepay register [contract_address] [wallet_limit]
```

> Note: the sender of this transaction must be the contract admin, if exists, or else the contract creator.

The `contract_address` is the bech32 address of the contract whose execution fees will be covered. The `wallet_limit` is the maximum number of times a wallet can execute the contract with 0 fees. This is a safety measure to prevent draining the account. The `wallet_limit` can be set to 0 to disable all FeePay interactions with this contract. Executions can still take place if the client explicitly specifies gas or a fee.

## Updating the Wallet Limit

The `wallet_limit` can be updated by executing the following transaction:

```bash
junod tx feepay update-wallet-limit [contract_address] [wallet_limit]
```

> Note: the sender of this transaction must be the contract admin, if exists, or else the contract creator.

The `contract_address` is the bech32 address of the FeePay contract to update. The `wallet_limit` is the maximum number of times a wallet can execute the contract with 0 fees. A `wallet_limit` of 0 disables all FeePay interactions with this contract. Executions can still take place if the client explicitly specifies gas or a fee.

## Unregistering a Contract

A contract can be unregistered by executing the following transaction:

```bash
junod tx feepay unregister [contract_address]
```

> Note: the sender of this transaction must be the contract admin, if exists, or else the contract creator.

The `contract_address` is the bech32 address of the FeePay contract to unregister. Unregistering a contract will remove it from the FeePay module. This means that clients will no longer be able to interact with the contract with 0 fees. Executions can still take place if the client explicitly specifies gas or a fee. All funds in the contract will be sent to the contract creator.

## Funding a Contract

A contract can be funded by executing the following transaction:

```bash
junod tx feepay fund [contract_address] [amount]
```

The `contract_address` is the bech32 address of the FeePay contract to fund. The `amount` is the amount of Juno to send to the contract. This amount will be used to pay for the execution fees of transactions interacting with the contract. Presently the FeePay module only supports Juno. The `amount` must be specified in ujuno.

## Client Interactions

Clients can interact with a contract registered with FeePay by explicitly specifying 0 fees. This can be done by setting the `--fees=0ujuno` flag.

```bash
junod tx wasm execute [contract_address] [json] --fees=0ujuno
```

The `contract_address` is the bech32 address of the FeePay contract to interact with. The `json` is the JSON-encoded transaction message. The `--fees=0ujuno` flag explicitly sets the fees to 0. This will trigger the FeePay module to attempt to pay for the execution fees of the transaction. See the [Ante](03_ante.md) for more details on how this works.
