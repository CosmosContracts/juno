# Clients

## Command Line Interface

Find below a list of `junod` commands added with the `x/cw-hooks` module. You can obtain the full list by using the `junod -h` command. A CLI command can look like this:

```bash
junod query cw-hooks params
```

### Queries

| Command            | Subcommand             | Description                              |
| :----------------- | :--------------------- | :--------------------------------------- |
| `query` `cw-hooks` | `params`               | Get module params                        |
| `query` `cw-hooks` | `governance-contracts` | Get registered governance contracts      |
| `query` `cw-hooks` | `staking-contracts`    | Get registered staking contracts         |

### Transactions

| Command         | Subcommand   | Description                           |
| :-------------- | :----------- | :------------------------------------ |
| `tx` `cw-hooks` | `register`   | Register a contract for events        |
| `tx` `cw-hooks` | `unregister` | Unregister a contract from events     |

## gRPC Queries

| Verb   | Method                                            |
| :----- | :------------------------------------------------ |
| `gRPC` | `juno.cwhooks.v1.Query/Params`                    |
| `gRPC` | `juno.cwhooks.v1.Query/StakingContracts`          |
| `gRPC` | `juno.cwhooks.v1.Query/GovernanceContracts`       |
| `GET`  | `/juno/cwhooks/v1/params`                         |
| `GET`  | `/juno/cwhooks/v1/staking_contracts`              |
| `GET`  | `/juno/cwhooks/v1/governance_contracts`           |

### gRPC Transactions

| Verb   | Method                                      |
| :----- | :------------------------------------------ |
| `gRPC` | `juno.cwhooks.v1.Msg/RegisterStaking`       |
| `gRPC` | `juno.cwhooks.v1.Msg/UnregisterStaking`     |
| `gRPC` | `juno.cwhooks.v1.Msg/RegisterGovernance`    |
| `gRPC` | `juno.cwhooks.v1.Msg/UnregisterGovernance`  |
| `POST` | `/juno/cwhooks/v1/tx/register_staking`      |
| `POST` | `/juno/cwhooks/v1/tx/unregister_staking`    |
| `POST` | `/juno/cwhooks/v1/tx/register_governance`   |
| `POST` | `/juno/cwhooks/v1/tx/unregister_governance` |
