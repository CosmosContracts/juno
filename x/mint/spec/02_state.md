<!--
order: 2
-->

# State

## Minter

The minter is a space for holding current inflation information.

- Minter: `0x00 -> ProtocolBuffer(minter)`

```go
type Minter struct {
 Inflation        sdk.Dec   // current annual inflation rate
 Phase            uint64    // current phase inflation
 StartPhaseBlock  uint64    // current phase start block 
 AnnualProvisions sdk.Dec   // current annual exptected provisions
}
```

## Params

Minting params are held in the global params store.

- Params: `mint/params -> legacy_amino(params)`

```go
type Params struct {
 MintDenom           string  // type of coin to mint
 BlocksPerYear       uint64   // expected blocks per year
}
```
