# `cw-hooks`

## Abstract

This document specifies the internal `x/cw-hooks` module of Juno Network.

The `x/cw-hooks` module allows specific contracts to be executed at the end of every block. This allows the smart contract to perform actions that may need to happen every block, or at set block intervals.

By using this module, your application can remove the headache of external whitelisted bots and instead depend on the chain itself for constant executions.

## Contents

1. **[Concepts](./00_concepts.md)**
1. **[Clients](./01_clients.md)**
1. **[State](./02_state.md)**
1. **[Params](./03_params.md)**
1. **[Register Contracts](./04_register.md)**
1. **[Unregister Contracts](./05_stop_events.md)**
1. **[Rust CW Integration](./06_integration.md)**
