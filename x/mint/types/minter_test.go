package types

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestPhaseInflation(t *testing.T) {
	minter := DefaultInitialMinter()

	tests := []struct {
		phase        uint64
		expInflation sdkmath.LegacyDec
	}{
		// phase 1, inflation: 40%
		{1, sdkmath.LegacyNewDecWithPrec(40, 2)},
		// phase 2, inflation: 20%
		{2, sdkmath.LegacyNewDecWithPrec(20, 2)},
		// phase 3, inflation: 10%
		{3, sdkmath.LegacyNewDecWithPrec(10, 2)},
		// phase 4, inflation: 9%
		{4, sdkmath.LegacyNewDecWithPrec(9, 2)},
		// phase 5, inflation: 8%
		{5, sdkmath.LegacyNewDecWithPrec(8, 2)},
		// phase 6, inflation: 7%
		{6, sdkmath.LegacyNewDecWithPrec(7, 2)},
		// phase 7, inflation: 6%
		{7, sdkmath.LegacyNewDecWithPrec(6, 2)},
		// phase 8, inflation: 5%
		{8, sdkmath.LegacyNewDecWithPrec(5, 2)},
		// phase 9, inflation: 4%
		{9, sdkmath.LegacyNewDecWithPrec(4, 2)},
		// phase 10, inflation: 3%
		{10, sdkmath.LegacyNewDecWithPrec(3, 2)},
		// phase 11, inflation: 2%
		{11, sdkmath.LegacyNewDecWithPrec(2, 2)},
		// phase 12, inflation: 1%
		{12, sdkmath.LegacyNewDecWithPrec(1, 2)},
		// end phase, inflation: 0%
		{13, sdkmath.LegacyNewDecWithPrec(0, 2)},
		// end phase, inflation: 0%
		{23, sdkmath.LegacyNewDecWithPrec(0, 2)},
	}
	for i, tc := range tests {
		inflation := minter.PhaseInflationRate(tc.phase)

		require.True(t, inflation.Equal(tc.expInflation),
			"Test Index: %v\nInflation:  %v\nExpected: %v\n", i, inflation, tc.expInflation)
	}
}

func TestNextPhase(t *testing.T) {
	minter := DefaultInitialMinter()
	params := DefaultParams()

	blocksPerYear := uint64(100)
	tests := []struct {
		currentBlock, currentPhase, startPhaseBlock, blocksYear, expPhase uint64
		currentSupply                                                     sdkmath.Int
		targetSupply                                                      sdkmath.Int
	}{
		{1, 0, 0, blocksPerYear, 1, sdkmath.NewInt(10000), sdkmath.NewInt(14000)},
		{50, 1, 1, blocksPerYear, 1, sdkmath.NewInt(12000), sdkmath.NewInt(14000)},
		// if targetSupply is > currentSupply it doesn't
		// matter how much by
		{99, 1, 1, blocksPerYear, 1, sdkmath.NewInt(13960), sdkmath.NewInt(1140000)},
		{100, 1, 1, blocksPerYear, 2, sdkmath.NewInt(14000), sdkmath.NewInt(14000)},
		{101, 1, 1, blocksPerYear, 2, sdkmath.NewInt(16000), sdkmath.NewInt(14000)},
		// since currentSupply is larger than targetSupply
		// next phase returns phase + 1 regardless of inputs
		{102, 2, 101, blocksPerYear, 3, sdkmath.NewInt(29000), sdkmath.NewInt(14000)},
	}
	for i, tc := range tests {
		minter.Phase = tc.currentPhase
		minter.StartPhaseBlock = tc.startPhaseBlock
		minter.TargetSupply = tc.targetSupply
		params.BlocksPerYear = tc.blocksYear

		phase := minter.NextPhase(params, tc.currentSupply)

		require.True(t, phase == tc.expPhase,
			"Test Index: %v\nPhase:  %v\nExpected: %v\n", i, phase, tc.expPhase)
	}
}

func TestBlockProvision(t *testing.T) {
	minter := InitialMinter(sdkmath.LegacyNewDecWithPrec(1, 1))
	params := DefaultParams()

	secondsPerYear := int64(60 * 60 * 8766)

	tests := []struct {
		annualProvisions int64
		expProvisions    int64
		totalSupply      sdkmath.Int
	}{
		{secondsPerYear / 5, 1, sdkmath.NewInt(1)},
		{secondsPerYear/5 + 1, 1, sdkmath.NewInt(1)},
		{(secondsPerYear / 5) * 2, 2, sdkmath.NewInt(1)},
		{(secondsPerYear / 5) / 2, 0, sdkmath.NewInt(1)},
		{(secondsPerYear / 5) * 3, 3, sdkmath.NewInt(1)},
		{(secondsPerYear / 5) * 7, 7, sdkmath.NewInt(2)},
		// we special case this below to trigger the
		// conditional in BlockProvision
		{(secondsPerYear / 5) * 7200, 0, sdkmath.NewInt(7000)},
	}
	for i, tc := range tests {
		minter.AnnualProvisions = sdkmath.LegacyNewDec(tc.annualProvisions)

		// if provision amount + total current supply
		// (totalSupply) exceeds target supply it should
		// return targetSupply - totalSupply, i.e. zero
		if i == 6 {
			minter.TargetSupply = tc.totalSupply
		} else {
			minter.TargetSupply = tc.totalSupply.Add(minter.AnnualProvisions.TruncateInt())
		}

		provisions := minter.BlockProvision(params, tc.totalSupply)

		expProvisions := sdk.NewCoin(params.MintDenom,
			sdkmath.NewInt(tc.expProvisions))

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
	b.ReportAllocs()
	minter := InitialMinter(sdkmath.LegacyNewDecWithPrec(1, 1))
	params := DefaultParams()

	s1 := rand.NewSource(100)
	r1 := rand.New(s1)
	minter.AnnualProvisions = sdkmath.LegacyNewDec(r1.Int63n(1000000))
	totalSupply := sdkmath.NewInt(100000000000000)
	minter.TargetSupply = sdkmath.NewInt(200000000000000)

	// run the BlockProvision function b.N times
	for n := 0; n < b.N; n++ {
		minter.BlockProvision(params, totalSupply)
	}
}

// Next inflation benchmarking
// BenchmarkPhaseInflation-4 1000000 1828 ns/op
func BenchmarkPhaseInflation(b *testing.B) {
	b.ReportAllocs()
	minter := InitialMinter(sdkmath.LegacyNewDecWithPrec(1, 1))
	phase := uint64(4)

	// run the PhaseInflationRate function b.N times
	for n := 0; n < b.N; n++ {
		minter.PhaseInflationRate(phase)
	}
}

// Next annual provisions benchmarking
// BenchmarkNextAnnualProvisions-4 5000000 251 ns/op
func BenchmarkNextAnnualProvisions(b *testing.B) {
	b.ReportAllocs()
	minter := InitialMinter(sdkmath.LegacyNewDecWithPrec(1, 1))
	params := DefaultParams()
	totalSupply := sdkmath.NewInt(100000000000000)

	// run the NextAnnualProvisions function b.N times
	for n := 0; n < b.N; n++ {
		minter.NextAnnualProvisions(params, totalSupply)
	}
}
