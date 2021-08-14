<!--
order: 2
-->

# State

## Minter

The minter is a space for holding current inflation information.

 - Minter: `0x00 -> amino(minter)`

```go
type Minter struct {
	Inflation        sdk.Dec   // current annual inflation rate
	AnnualProvisions sdk.Dec   // current annual exptected provisions
}
```

## Params

Minting params are held in the global params store. 

 - Params: `mint/params -> amino(params)`

```go
type Params struct {
	MintDenom           string  // type of coin to mint
	BlocksPerYear       uint64   // expected blocks per year
}
```
