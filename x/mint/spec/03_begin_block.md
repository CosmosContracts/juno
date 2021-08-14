<!--
order: 3
-->

# Begin-Block

Minting parameters are recalculated and inflation
paid at the beginning of each block.

## NextInflationRate

The target annual inflation rate is recalculated each block and stored if it changes (new phase)

```
func (m Minter) NextInflationRate(params Params, currentBlock sdk.Dec) sdk.Dec {
	phase := currentBlock.Quo(sdk.NewDec(int64(params.BlocksPerYear))).Ceil()

	switch {
	case phase.GT(sdk.NewDec(12)):
		return sdk.ZeroDec()

	case phase.Equal(sdk.NewDec(1)):
		return sdk.NewDecWithPrec(40, 2)

	case phase.Equal(sdk.NewDec(2)):
		return sdk.NewDecWithPrec(20, 2)

	case phase.Equal(sdk.NewDec(3)):
		return sdk.NewDecWithPrec(10, 2)

	default:
		return sdk.NewDecWithPrec(13-phase.RoundInt64(), 2)
	}
}
```

## NextAnnualProvisions

Calculate the annual provisions based on current total supply and inflation
rate. This parameter is calculated once per new inflation rate. 

```
NextAnnualProvisions(params Params, totalSupply sdk.Dec) (provisions sdk.Dec) {
	return Inflation * totalSupply
```

## BlockProvision

Calculate the provisions generated for each block based on current annual provisions. The provisions are then minted by the `mint` module's `ModuleMinterAccount` and then transferred to the `auth`'s `FeeCollector` `ModuleAccount`.

```
BlockProvision(params Params) sdk.Coin {
	provisionAmt = AnnualProvisions/ params.BlocksPerYear
	return sdk.NewCoin(params.MintDenom, provisionAmt.Truncate())
```
