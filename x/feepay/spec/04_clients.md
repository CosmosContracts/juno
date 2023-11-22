<!--
order: 4
-->

# Clients

## Command Line Interface (CLI)

The CLI has been updated with new queries and transactions for the `x/feepay` module. View the entire list below.

### Queries

| Command              | Subcommand    | Arguments                           | Description                                                     |
| :------------------- | :------------ | :---------------------------------- | :-------------------------------------------------------------- |
| `junod query feepay` | `params`      |                                     | Get FeePay params                                               |
| `junod query feepay` | `contract`    | [contract_address]                  | Get a FeePay contract                                           |
| `junod query feepay` | `contracts`   |                                     | Get all FeePay contracts                                        |
| `junod query feepay` | `uses`        | [contract_address] [wallet_address] | Get the number of times a wallet has interacted with a contract |
| `junod query feepay` | `is-eligible` | [contract_address] [wallet_address] | Check if a wallet has not met the wallet limit on a contract    |

### Transactions

| Command           | Subcommand            | Arguments                         | Description                                    |
| :-----------------| :-------------------- | :-------------------------------- | :--------------------------------------------- |
| `junod tx feepay` | `register`            | [contract_address] [wallet_limit] | Register a FeePay contract with a wallet limit |
| `junod tx feepay` | `update-wallet-limit` | [contract_address] [wallet_limit] | Update the wallet limit of a FeePay contract   |
| `junod tx feepay` | `unregister`          | [contract_address]                | Unregister a FeePay contract                   |
| `junod tx feepay` | `fund`                | [contract_address] [amount]       | Fund a FeePay contract                         |
