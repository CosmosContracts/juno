package types

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestNextInflation(t *testing.T) {
	minter := DefaultInitialMinter()
	params := DefaultParams()
	blocksPerYr := sdk.NewDec(int64(params.BlocksPerYear))

	// Governing Mechanism:
<<<<<<< HEAD
	//    inflationRateChangePerYear = (1- BondedRatio/ GoalBonded) * MaxInflationRateChange

	tests := []struct {
		bondedRatio, setInflation, expChange sdk.Dec
	}{
		// with 0% bonded atom supply the inflation should increase by InflationRateChange
		{sdk.ZeroDec(), sdk.NewDecWithPrec(7, 2), params.InflationRateChange.Quo(blocksPerYr)},

		// 100% bonded, starting at 20% inflation and being reduced
		// (1 - (1/0.67))*(0.13/8667)
		{sdk.OneDec(), sdk.NewDecWithPrec(20, 2),
			sdk.OneDec().Sub(sdk.OneDec().Quo(params.GoalBonded)).Mul(params.InflationRateChange).Quo(blocksPerYr)},

		// 50% bonded, starting at 10% inflation and being increased
		{sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(10, 2),
			sdk.OneDec().Sub(sdk.NewDecWithPrec(5, 1).Quo(params.GoalBonded)).Mul(params.InflationRateChange).Quo(blocksPerYr)},

		// test 7% minimum stop (testing with 100% bonded)
		{sdk.OneDec(), sdk.NewDecWithPrec(7, 2), sdk.ZeroDec()},
		{sdk.OneDec(), sdk.NewDecWithPrec(700000001, 10), sdk.NewDecWithPrec(-1, 10)},

		// test 20% maximum stop (testing with 0% bonded)
		{sdk.ZeroDec(), sdk.NewDecWithPrec(20, 2), sdk.ZeroDec()},
		{sdk.ZeroDec(), sdk.NewDecWithPrec(1999999999, 10), sdk.NewDecWithPrec(1, 10)},

		// perfect balance shouldn't change inflation
		{sdk.NewDecWithPrec(67, 2), sdk.NewDecWithPrec(15, 2), sdk.ZeroDec()},
	}
	for i, tc := range tests {
		minter.Inflation = tc.setInflation

		inflation := minter.NextInflationRate(params, tc.bondedRatio)
		diffInflation := inflation.Sub(tc.setInflation)

		require.True(t, diffInflation.Equal(tc.expChange),
			"Test Index: %v\nDiff:  %v\nExpected: %v\n", i, diffInflation, tc.expChange)
=======
	//    Juno tokenomics

	firstBlockInYear := func(year int64) sdk.Dec {
		return blocksPerYr.Mul(sdk.NewDec(year - 1)).Add(sdk.OneDec())
	}

	tests := []struct {
		currentBlock, expInflation sdk.Dec
	}{
		// phase 1, inflation: 40%
		{firstBlockInYear(1), sdk.NewDecWithPrec(40, 2)},
		// phase 2, inflation: 20%
		{firstBlockInYear(2), sdk.NewDecWithPrec(20, 2)},
		// phase 3, inflation: 10%
		{firstBlockInYear(3), sdk.NewDecWithPrec(10, 2)},
		// phase 4, inflation: 9%
		{firstBlockInYear(4), sdk.NewDecWithPrec(9, 2)},
		// phase 5, inflation: 8%
		{firstBlockInYear(5), sdk.NewDecWithPrec(8, 2)},
		// phase 6, inflation: 7%
		{firstBlockInYear(6), sdk.NewDecWithPrec(7, 2)},
		// phase 7, inflation: 6%
		{firstBlockInYear(7), sdk.NewDecWithPrec(6, 2)},
		// phase 8, inflation: 5%
		{firstBlockInYear(8), sdk.NewDecWithPrec(5, 2)},
		// phase 9, inflation: 4%
		{firstBlockInYear(9), sdk.NewDecWithPrec(4, 2)},
		// phase 10, inflation: 3%
		{firstBlockInYear(10), sdk.NewDecWithPrec(3, 2)},
		// phase 11, inflation: 2%
		{firstBlockInYear(11), sdk.NewDecWithPrec(2, 2)},
		// phase 12, inflation: 1%
		{firstBlockInYear(12), sdk.NewDecWithPrec(1, 2)},
		// end phase, inflation: 0%
		{firstBlockInYear(13), sdk.NewDecWithPrec(0, 2)},
		// end phase, inflation: 0%
		{firstBlockInYear(23), sdk.NewDecWithPrec(0, 2)},
	}
	for i, tc := range tests {
		inflation := minter.NextInflationRate(params, tc.currentBlock)

		require.True(t, inflation.Equal(tc.expInflation),
			"Test Index: %v\nInflation:  %v\nExpected: %v\n", i, inflation, tc.expInflation)
>>>>>>> disperze/mint-module
	}
}

func TestBlockProvision(t *testing.T) {
	minter := InitialMinter(sdk.NewDecWithPrec(1, 1))
	params := DefaultParams()

	secondsPerYear := int64(60 * 60 * 8766)

	tests := []struct {
		annualProvisions int64
		expProvisions    int64
	}{
		{secondsPerYear / 5, 1},
		{secondsPerYear/5 + 1, 1},
		{(secondsPerYear / 5) * 2, 2},
		{(secondsPerYear / 5) / 2, 0},
	}
	for i, tc := range tests {
		minter.AnnualProvisions = sdk.NewDec(tc.annualProvisions)
		provisions := minter.BlockProvision(params)

		expProvisions := sdk.NewCoin(params.MintDenom,
			sdk.NewInt(tc.expProvisions))

		require.True(t, expProvisions.IsEqual(provisions),
			"test: %v\n\tExp: %v\n\tGot: %v\n",
			i, tc.expProvisions, provisions)
	}
}

// Benchmarking :)
// previously using sdk.Int operations:
// BenchmarkBlockProvision-4 5000000 220 ns/op
//
// using sdk.Dec operations: (current implementation)
// BenchmarkBlockProvision-4 3000000 429 ns/op
func BenchmarkBlockProvision(b *testing.B) {
	minter := InitialMinter(sdk.NewDecWithPrec(1, 1))
	params := DefaultParams()

	s1 := rand.NewSource(100)
	r1 := rand.New(s1)
	minter.AnnualProvisions = sdk.NewDec(r1.Int63n(1000000))

	// run the BlockProvision function b.N times
	for n := 0; n < b.N; n++ {
		minter.BlockProvision(params)
	}
}

// Next inflation benchmarking
// BenchmarkNextInflation-4 1000000 1828 ns/op
func BenchmarkNextInflation(b *testing.B) {
	minter := InitialMinter(sdk.NewDecWithPrec(1, 1))
	params := DefaultParams()
<<<<<<< HEAD
	bondedRatio := sdk.NewDecWithPrec(1, 1)

	// run the NextInflationRate function b.N times
	for n := 0; n < b.N; n++ {
		minter.NextInflationRate(params, bondedRatio)
=======
	currentBlock := sdk.NewDec(1)

	// run the NextInflationRate function b.N times
	for n := 0; n < b.N; n++ {
		minter.NextInflationRate(params, currentBlock)
>>>>>>> disperze/mint-module
	}

}

// Next annual provisions benchmarking
// BenchmarkNextAnnualProvisions-4 5000000 251 ns/op
func BenchmarkNextAnnualProvisions(b *testing.B) {
	minter := InitialMinter(sdk.NewDecWithPrec(1, 1))
	params := DefaultParams()
	totalSupply := sdk.NewInt(100000000000000)

	// run the NextAnnualProvisions function b.N times
	for n := 0; n < b.N; n++ {
		minter.NextAnnualProvisions(params, totalSupply)
	}

}
