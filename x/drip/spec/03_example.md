<!--
order: 3
-->

# Example contract

The following code is an example function to encode a `MsgDistributeTokens` message in CosmWasm

```rust
fn encode_msg_create_vesting_acct(vest_to: &Addr, env: Env) -> Result<Response, ContractError> {

    // Test with 1 ucosm
    let one_cosm: Coin = coin(1_000_000, "ucosm");

    // MsgDistributeTokens

    let proto = Anybuf::new()
        .append_string(1, &env.contract.address)
        .append_message(2, &Anybuf::new()
            .append_string(1, &one_cosm.denom)
            .append_string(2, &one_cosm.amount.to_string())
        )
        .into_vec();

    let msg = CosmosMsg::Stargate { 
        type_url: "/juno.drip.v1.MsgDistributeTokens".to_string(), 
        value: proto.into() 
    };

}
```

More information about [https://lib.rs/crates/anybuf](Anybuf)