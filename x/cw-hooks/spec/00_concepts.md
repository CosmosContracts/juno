# Concepts

## CW-Hooks

The CW-Hooks module (CosmWasm Hooks) is an event based module which allowed developers to subscribe to wallet specific actions on Juno. This allows developers to write applications which need to following staking or governance actions for any account who performs them.

Some Examples could include:

Gov:

- When an account votes, send them 1 of your token-factory token.
- When a new proposal on the main chain is created, update your contract to allow for secondary voting through a contract or DAO.

Staking:

- When a user delegates tokens to a validator, update their new balance in the contract. This allows for synthetic DAO voting power relative to their chain stake
- Before and After a validator is slashed, auto update a vesting contract to slash the receiver of funds.
- When a validator gets into the active set, update your contract to remove the old validator.

## Registration

Developers register their contract(s) to receive fire-and-forget messages from the CW-Hooks module. This allows developers to write applications which need to following staking or governance actions for any account who performs them. Including standard wallets, DAOs, and other contracts.

### Limitations

By default, your contract can only perform 250,000 Gas execution per event. This is to prevent malicious contracts from spamming the network since all executes are feeless. If you need to perform more than 250,000 Gas execution, you can submit a proposal to increase this.
