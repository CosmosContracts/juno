# Retired globalfee codec compatibility

The globalfee keeper, stores, module manager registration, and transaction routing remain removed. This directory preserves only the historical generated message/parameter types and codec registration needed to decode governance proposals retained in mainnet state.

Removing these types makes `junod export` and historical governance queries panic on `/gaia.globalfee.v1beta1.MsgUpdateParams`.

The restored files come from this repository immediately before commit `0c859ad0` removed the module.
