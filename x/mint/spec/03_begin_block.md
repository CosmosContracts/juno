<!--
order: 3
-->

# Begin-Block

Minting parameters are recalculated and inflation
paid at the beginning of each block.

## PhaseInflationRate

The target annual inflation rate is recalculated each block and stored if it changes (new phase)

```go
func (m Minter) PhaseInflationRate(phase uint64) sdk.Dec {
 switch {
 case phase > 12:
  return sdkmath.LegacyZeroDec()

 case phase == 1:
  return sdkmath.LegacyNewDecWithPrec(40, 2)

 case phase == 2:
  return sdkmath.LegacyNewDecWithPrec(20, 2)

 case phase == 3:
  return sdkmath.LegacyNewDecWithPrec(10, 2)

 default:
  return sdkmath.LegacyNewDecWithPrec(13-int64(phase), 2)
 }
}
```

## NextAnnualProvisions

Calculate the annual provisions based on current total supply and inflation
rate. This parameter is calculated once per new inflation rate.

```go
NextAnnualProvisions(params Params, totalSupply sdkmath.LegacyDec) (provisions sdkmath.LegacyDec) {
 return Inflation * totalSupply
```

## BlockProvision

Calculate the provisions generated for each block based on current annual provisions. The provisions are then minted by the `mint` module's `ModuleMinterAccount` and then transferred to the `auth`'s `FeeCollector` `ModuleAccount`.

```go
BlockProvision(params Params) sdk.Coin {
 provisionAmt = AnnualProvisions / params.BlocksPerYear
 return sdk.NewCoin(params.MintDenom, provisionAmt.Truncate())
```
