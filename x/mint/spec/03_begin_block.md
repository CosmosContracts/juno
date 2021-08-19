<!--
order: 3
-->

# Begin-Block

Minting parameters are recalculated and inflation
paid at the beginning of each block.

## NextInflationRate

<<<<<<< HEAD
The target annual inflation rate is recalculated each block.
The inflation is also subject to a rate change (positive or negative)
depending on the distance from the desired ratio (67%). The maximum rate change
possible is defined to be 13% per year, however the annual inflation is capped
as between 7% and 20%.

```
NextInflationRate(params Params, bondedRatio sdk.Dec) (inflation sdk.Dec) {
	inflationRateChangePerYear = (1 - bondedRatio/params.GoalBonded) * params.InflationRateChange
	inflationRateChange = inflationRateChangePerYear/blocksPerYr

	// increase the new annual inflation for this next cycle
	inflation += inflationRateChange
	if inflation > params.InflationMax {
		inflation = params.InflationMax
	}
	if inflation < params.InflationMin {
		inflation = params.InflationMin
	}

	return inflation
=======
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
>>>>>>> disperze/mint-module
}
```

## NextAnnualProvisions

Calculate the annual provisions based on current total supply and inflation
<<<<<<< HEAD
rate. This parameter is calculated once per block. 
=======
rate. This parameter is calculated once per new inflation rate. 
>>>>>>> disperze/mint-module

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
