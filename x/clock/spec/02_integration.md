# CosmWasm Integration

This module does not require any custom bindings. Rather, you must just add the following Sudo message to your contract. If your contract is not whitelisted, you can still upload it to the chain. However, to get it to execute, you must submit a proposal to add your contract to the whitelist.

You can find a basic [cw-clock contract here](https://github.com/Reecepbcups/cw-clock-example)

## Implementation

Add the following to your Rust Contract:

```rust
// msg.rs
#[cw_serde]
pub enum SudoMsg {    
    ClockEndBlock { },
}

// contract.rs
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn sudo(deps: DepsMut, _env: Env, msg: SudoMsg) -> Result<Response, ContractError> {
    match msg {        
        SudoMsg::ClockEndBlock { } => {
            let mut config = CONFIG.load(deps.storage)?;
            config.val += 1;
            CONFIG.save(deps.storage, &config)?;

            Ok(Response::new())
        }
    }
}
```

Using the above example, for every block the module will increase the `val` Config variable by 1. This is a simple example, but you can use this to perform any action you want (ex: cleanup, auto compounding, etc).

If you wish not to have your action performed every block, you can use the `env` variable in the Sudo message to check the block height and only perform an action every X blocks.

```rust
// contract.rs
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn sudo(deps: DepsMut, env: Env, msg: SudoMsg) -> Result<Response, ContractError> {
    match msg {        
        SudoMsg::ClockEndBlock { } => {    
            // If the block is not divisible by ten, do nothing.      
            if env.block.height % 10 != 0 {
                return Ok(Response::new());
            }

            let mut config = CONFIG.load(deps.storage)?;
            config.val += 1;
            CONFIG.save(deps.storage, &config)?;

            Ok(Response::new())
        }
    }
}
```
