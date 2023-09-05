# CosmWasm Integration

This module does not require any custom bindings. Rather, you must just add the following Sudo messages to your contract. Any contract is able to register itself via the following command

```bash
junod tx register juno-hooks
```

You can find a basic [juno-staking-hooks contract here](https://github.com/Reecepbcups/cw-juno-staking-hooks-example)

## Implementation

Add the following to your Rust Contract:

```rust
// msg.rs
#[cw_serde]
pub enum SudoMsg {    
    ...
}

// contract.rs
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn sudo(deps: DepsMut, _env: Env, msg: SudoMsg) -> Result<Response, ContractError> {
    match msg {        
        ...
    }
}
```
