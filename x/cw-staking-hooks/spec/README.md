# `clock`

## Abstract

This document specifies the internal `x/cw-staking-hooks` module of Juno Network.

The `x/cw-staking-hooks` module allows specific contracts to be executed at the end of every block. This allows the smart contract to perform actions that may need to happen every block, or at set block intervals.

By using this module, your application can remove the headache of external whitelisted bots and instead depend on the chain itself for constant executions.

## Contents

1. **[Authorization](01_authorization.md)**
2. **[Contract Integration](02_integration.md)**
