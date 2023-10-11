# Register a contract

See [Integration](./06_integration.md)) for all events emitted by the chain.

## Staking Events

> `junod tx cw-hooks register staking [contract_bech32] --from [admin|creator]`

*Registers the contract to receive staking events (fire and forget)*

## Governance Events

> `junod tx cw-hooks register governance [contract_bech32] --from [admin|creator]`

*Registers the contract to receive governance events (fire and forget)*

---

### Parameters

`contract_bech32 (string, required)`: The bech32 address of the contract who will receive the updates.

### Permissions

This command can only be run by the admin of the contract. If there is no admin, then it can only be run by the contract creator.
