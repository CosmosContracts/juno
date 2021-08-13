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
	//    Juno tokenomics

	tests := []struct {
		currentBlock, expInflation sdk.Dec
	}{
		// phase 1, inflation: 40%
		{sdk.OneDec(), sdk.NewDecWithPrec(40, 2)},
		// phase 2, inflation: 20%
		{blocksPerYr.Add(sdk.OneDec()), sdk.NewDecWithPrec(20, 2)},
		// phase 3, inflation: 10%
		{blocksPerYr.Mul(sdk.NewDec(2)).Add(sdk.OneDec()), sdk.NewDecWithPrec(10, 2)},
		// phase 4, inflation: 9%
		{blocksPerYr.Mul(sdk.NewDec(3)).Add(sdk.OneDec()), sdk.NewDecWithPrec(9, 2)},
		// phase 5, inflation: 8%
		{blocksPerYr.Mul(sdk.NewDec(4)).Add(sdk.OneDec()), sdk.NewDecWithPrec(8, 2)},
		// phase 6, inflation: 7%
		{blocksPerYr.Mul(sdk.NewDec(5)).Add(sdk.OneDec()), sdk.NewDecWithPrec(7, 2)},
		// phase 7, inflation: 6%
		{blocksPerYr.Mul(sdk.NewDec(6)).Add(sdk.OneDec()), sdk.NewDecWithPrec(6, 2)},
		// phase 8, inflation: 5%
		{blocksPerYr.Mul(sdk.NewDec(7)).Add(sdk.OneDec()), sdk.NewDecWithPrec(5, 2)},
		// phase 9, inflation: 4%
		{blocksPerYr.Mul(sdk.NewDec(8)).Add(sdk.OneDec()), sdk.NewDecWithPrec(4, 2)},
		// phase 10, inflation: 3%
		{blocksPerYr.Mul(sdk.NewDec(9)).Add(sdk.OneDec()), sdk.NewDecWithPrec(3, 2)},
		// phase 11, inflation: 2%
		{blocksPerYr.Mul(sdk.NewDec(10)).Add(sdk.OneDec()), sdk.NewDecWithPrec(2, 2)},
		// phase 12, inflation: 1%
		{blocksPerYr.Mul(sdk.NewDec(11)).Add(sdk.OneDec()), sdk.NewDecWithPrec(1, 2)},
		// end phase, inflation: 0%
		{blocksPerYr.Mul(sdk.NewDec(12)).Add(sdk.OneDec()), sdk.NewDecWithPrec(0, 2)},
	}
	for i, tc := range tests {
		inflation := minter.NextInflationRate(params, tc.currentBlock)

		require.True(t, inflation.Equal(tc.expInflation),
			"Test Index: %v\nInflation:  %v\nExpected: %v\n", i, inflation, tc.expInflation)
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
	currentBlock := sdk.NewDec(1)

	// run the NextInflationRate function b.N times
	for n := 0; n < b.N; n++ {
		minter.NextInflationRate(params, currentBlock)
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
