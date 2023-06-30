# ROADMAP

This document contains the roadmap for the Juno project. It is a living document and will be updated as the project progresses. For the most update to date information, please follow the Notion and tracking issue links in each section.

---

## Long Term - Q4 2023+

- [Long term Tracking Issue](https://github.com/CosmosContracts/juno/issues/611)
- SDK v0.47, Tendermint 0.37
- IBC v5/6, ICA v3 (optional)
- [Native Liquid Staking](https://github.com/iqlusioninc/liquidity-staking-module)
- [Improving the Nakamoto Coefficient](https://github.com/CosmosContracts/juno/issues/474)
- Wasm based oracle

---

## V15 - End of Q2 2023

<!-- - [Medium Blog](...) -->

<!-- - [V15 Tracking Issue](...) -->

This upgrade focuses entirely on moving Juno's block times from 6 seconds to 3 seconds.

## Features

- 3 second block times (from 6 seconds)

---

## V14 (Aurora) - Early Q2 2023

- [Medium Blog](https://medium.com/@JunoNetwork/jun%C3%B8-aurora-ac67a8143e22)

- [V14 Tracking Issue](https://github.com/CosmosContracts/juno/issues/548)

This update will focus more on upgrading the base layer of the Juno stack, bringing new features and pushing us to the latest versions of the software.

## V14 Features

- IBC-Hooks
- CosmWasm v0.31
- WasmVM v1.2.1
- Global Minimum Fees (governance controlled)
- Feeless IBC Relaying
- TokenFactory: burnFrom, burnTo, ForceTransfer
- [Interchain test](https://github.com/strangelove-ventures/interchaintest)
- Using [Skip's MEV Tendermint fork](https://github.com/skip-mev/mev-cometbft) by default

---

## V13 - Q1 2023

Links:

- [Medium Blog](https://medium.com/@JunoNetwork/jun%C3%B8-v-13-fefa9d2dfce5)

- [v13 Tracking Issue](https://github.com/CosmosContracts/juno/issues/475)

The V13 update is Juno's largest update, bringing many new features for developers, users, and relayers.

### V13 PRs

- [x/FeeShare (CosmWasm)](https://github.com/CosmosContracts/juno/pull/385)
- [x/TokenFactory](https://github.com/CosmosContracts/juno/pull/368)
- [Packet Forward Middleware](https://github.com/CosmosContracts/juno/pull/513)
- [x/GlobalFee](https://github.com/CosmosContracts/juno/pull/411)
- [More ICA Messages](https://github.com/CosmosContracts/juno/pull/436/files)
- [Governance Spam Prevention](https://github.com/CosmosContracts/juno/pull/394)
- [x/wasmd 30](https://github.com/CosmosContracts/juno/pull/387)
- [x/ibc V4](https://github.com/CosmosContracts/juno/pull/387)
- [x/ibc-fees](https://github.com/CosmosContracts/juno/pull/432)

V13 is targeted at developers with relayer and user experience improvements as well.

**FeeShare** will allow contract developers to receive 50% of gas fees executed on their contract. Providing an alternative income source for new business use cases. This also enhances current business models to support developers & grow the ecosystem further.

The **TokenFactory** will make developers' lives easier, and also make querying users' [DAO](https://daodao.zone/) tokens via MintScan and Keplr possible. By default, CosmWasm smart contracts accept native tokens. However, the only initial native tokens are the staking demons for most chains. This gives the ability for a user to create their token, and manage the tokenomics behind it. Then accept it just as they would any other denomination via the standard [x/bank](https://github.com/cosmos/cosmos-sdk/tree/main/x/bank) module.

Speaking of relayers, **IBCFees** now helps to fund those who relayer your packets! In the above paragraph, we mention how IBC transfers are feeless for relayers. Fees can still be sent with these packets and bring some income for relayers, thus maintaining public goods infrastructure. The relayers still have to pay the fee on the other chain's token, but this is a positive step in the right direction for variables we can control.
