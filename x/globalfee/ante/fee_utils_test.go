package ante

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestContainZeroCoins(t *testing.T) {
	zeroCoin1 := sdk.NewCoin("photon", sdkmath.ZeroInt())
	zeroCoin2 := sdk.NewCoin("stake", sdkmath.ZeroInt())
	coin1 := sdk.NewCoin("photon", sdkmath.NewInt(1))
	coin2 := sdk.NewCoin("stake", sdkmath.NewInt(2))
	coin3 := sdk.NewCoin("quark", sdkmath.NewInt(3))
	// coins must be valid !!!
	coinsEmpty := sdk.Coins{}
	coinsNonEmpty := sdk.Coins{coin1, coin2}
	coinsCointainZero := sdk.Coins{coin1, zeroCoin2}
	coinsCointainTwoZero := sdk.Coins{zeroCoin1, zeroCoin2, coin3}
	coinsAllZero := sdk.Coins{zeroCoin1, zeroCoin2}

	tests := []struct {
		c  sdk.Coins
		ok bool
	}{
		{
			coinsEmpty,
			true,
		},
		{
			coinsNonEmpty,
			false,
		},
		{
			coinsCointainZero,
			true,
		},
		{
			coinsCointainTwoZero,
			true,
		},
		{
			coinsAllZero,
			true,
		},
	}

	for _, test := range tests {
		ok := ContainZeroCoins(test.c)
		require.Equal(t, test.ok, ok)
	}
}

// Note that in a real Gaia deployment all zero coins can be removed from minGasPrice.
// This sanitizing happens when the minGasPrice is set into the context.
// (see baseapp.SetMinGasPrices in gaia/cmd/root.go line 221)
func TestCombinedFeeRequirement(t *testing.T) {
	zeroCoin1 := sdk.NewCoin("photon", sdkmath.ZeroInt())
	zeroCoin2 := sdk.NewCoin("stake", sdkmath.ZeroInt())
	zeroCoin3 := sdk.NewCoin("quark", sdkmath.ZeroInt())
	coin1 := sdk.NewCoin("photon", sdkmath.NewInt(1))
	coin2 := sdk.NewCoin("stake", sdkmath.NewInt(2))
	coin1High := sdk.NewCoin("photon", sdkmath.NewInt(10))
	coin2High := sdk.NewCoin("stake", sdkmath.NewInt(20))
	coinNewDenom1 := sdk.NewCoin("Newphoton", sdkmath.NewInt(1))
	coinNewDenom2 := sdk.NewCoin("Newstake", sdkmath.NewInt(1))
	// coins must be valid !!! and sorted!!!
	coinsEmpty := sdk.Coins{}
	coinsNonEmpty := sdk.Coins{coin1, coin2}.Sort()
	coinsNonEmptyHigh := sdk.Coins{coin1High, coin2High}.Sort()
	coinsNonEmptyOneHigh := sdk.Coins{coin1High, coin2}.Sort()
	coinsNewDenom := sdk.Coins{coinNewDenom1, coinNewDenom2}.Sort()
	coinsNewOldDenom := sdk.Coins{coin1, coinNewDenom1}.Sort()
	coinsNewOldDenomHigh := sdk.Coins{coin1High, coinNewDenom1}.Sort()
	coinsCointainZero := sdk.Coins{coin1, zeroCoin2}.Sort()
	coinsCointainZeroNewDenom := sdk.Coins{coin1, zeroCoin3}.Sort()
	coinsAllZero := sdk.Coins{zeroCoin1, zeroCoin2}.Sort()
	tests := map[string]struct {
		cGlobal  sdk.Coins
		c        sdk.Coins
		combined sdk.Coins
	}{
		"global fee empty, min fee empty, combined fee empty": {
			cGlobal:  coinsEmpty,
			c:        coinsEmpty,
			combined: coinsEmpty,
		},
		"global fee empty, min fee nonempty, combined fee empty": {
			cGlobal:  coinsEmpty,
			c:        coinsNonEmpty,
			combined: coinsEmpty,
		},
		"global fee nonempty, min fee empty, combined fee = global fee": {
			cGlobal:  coinsNonEmpty,
			c:        coinsNonEmpty,
			combined: coinsNonEmpty,
		},
		"global fee and min fee have overlapping denom, min fees amounts are all higher": {
			cGlobal:  coinsNonEmpty,
			c:        coinsNonEmptyHigh,
			combined: coinsNonEmptyHigh,
		},
		"global fee and min fee have overlapping denom, one of min fees amounts is higher": {
			cGlobal:  coinsNonEmpty,
			c:        coinsNonEmptyOneHigh,
			combined: coinsNonEmptyOneHigh,
		},
		"global fee and min fee have no overlapping denom, combined fee = global fee": {
			cGlobal:  coinsNonEmpty,
			c:        coinsNewDenom,
			combined: coinsNonEmpty,
		},
		"global fees and min fees have partial overlapping denom, min fee amount <= global fee amount, combined fees = global fees": {
			cGlobal:  coinsNonEmpty,
			c:        coinsNewOldDenom,
			combined: coinsNonEmpty,
		},
		"global fees and min fees have partial overlapping denom, one min fee amount > global fee amount, combined fee = overlapping highest": {
			cGlobal:  coinsNonEmpty,
			c:        coinsNewOldDenomHigh,
			combined: sdk.Coins{coin1High, coin2},
		},
		"global fees have zero fees, min fees have overlapping non-zero fees, combined fees = overlapping highest": {
			cGlobal:  coinsCointainZero,
			c:        coinsNonEmpty,
			combined: sdk.Coins{coin1, coin2},
		},
		"global fees have zero fees, min fees have overlapping zero fees": {
			cGlobal:  coinsCointainZero,
			c:        coinsCointainZero,
			combined: coinsCointainZero,
		},
		"global fees have zero fees, min fees have non-overlapping zero fees": {
			cGlobal:  coinsCointainZero,
			c:        coinsCointainZeroNewDenom,
			combined: coinsCointainZero,
		},
		"global fees are all zero fees, min fees have overlapping zero fees": {
			cGlobal:  coinsAllZero,
			c:        coinsAllZero,
			combined: coinsAllZero,
		},
		"global fees are all zero fees, min fees have overlapping non-zero fees, combined fee = overlapping highest": {
			cGlobal:  coinsAllZero,
			c:        coinsCointainZeroNewDenom,
			combined: sdk.Coins{coin1, zeroCoin2},
		},
		"global fees are all zero fees, fees have one overlapping non-zero fee": {
			cGlobal:  coinsAllZero,
			c:        coinsCointainZero,
			combined: coinsCointainZero,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			allFees := CombinedFeeRequirement(test.cGlobal, test.c)
			require.Equal(t, test.combined, allFees)
		})
	}
}

func TestSplitCoinsByDenoms(t *testing.T) {
	zeroGlobalFeesDenom0 := map[string]bool{}
	zeroGlobalFeesDenom1 := map[string]bool{
		"uatom":  true,
		"photon": true,
	}
	zeroGlobalFeesDenom2 := map[string]bool{
		"uatom": true,
	}
	zeroGlobalFeesDenom3 := map[string]bool{
		"stake": true,
	}

	photon := sdk.NewCoin("photon", sdkmath.OneInt())
	uatom := sdk.NewCoin("uatom", sdkmath.OneInt())
	feeCoins := sdk.NewCoins(photon, uatom)

	tests := map[string]struct {
		feeCoins             sdk.Coins
		zeroGlobalFeesDenom  map[string]bool
		expectedNonZeroCoins sdk.Coins
		expectedZeroCoins    sdk.Coins
	}{
		"no zero coins in global fees": {
			feeCoins:             feeCoins,
			zeroGlobalFeesDenom:  zeroGlobalFeesDenom0,
			expectedNonZeroCoins: feeCoins,
			expectedZeroCoins:    sdk.Coins{},
		},
		"no split of fee coins": {
			feeCoins:             feeCoins,
			zeroGlobalFeesDenom:  zeroGlobalFeesDenom3,
			expectedNonZeroCoins: feeCoins,
			expectedZeroCoins:    sdk.Coins{},
		},
		"split the fee coins": {
			feeCoins:             feeCoins,
			zeroGlobalFeesDenom:  zeroGlobalFeesDenom2,
			expectedNonZeroCoins: sdk.NewCoins(photon),
			expectedZeroCoins:    sdk.NewCoins(uatom),
		},
		"remove all of the fee coins": {
			feeCoins:             feeCoins,
			zeroGlobalFeesDenom:  zeroGlobalFeesDenom1,
			expectedNonZeroCoins: sdk.Coins{},
			expectedZeroCoins:    feeCoins,
		},
		"fee coins are empty": {
			feeCoins:             sdk.Coins{},
			zeroGlobalFeesDenom:  zeroGlobalFeesDenom1,
			expectedNonZeroCoins: sdk.Coins{},
			expectedZeroCoins:    sdk.Coins{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			feeCoinsNoZeroDenoms, feeCoinsZeroDenoms := splitCoinsByDenoms(test.feeCoins, test.zeroGlobalFeesDenom)
			require.Equal(t, test.expectedNonZeroCoins, feeCoinsNoZeroDenoms)
			require.Equal(t, test.expectedZeroCoins, feeCoinsZeroDenoms)
		})
	}
}

func TestSplitGlobalFees(t *testing.T) {
	photon0 := sdk.NewCoin("photon", sdkmath.ZeroInt())
	uatom0 := sdk.NewCoin("uatom", sdkmath.ZeroInt())
	photon1 := sdk.NewCoin("photon", sdkmath.OneInt())
	uatom1 := sdk.NewCoin("uatom", sdkmath.OneInt())

	globalFeesEmpty := sdk.Coins{}
	globalFees := sdk.Coins{photon1, uatom1}.Sort()
	globalFeesZeroCoins := sdk.Coins{photon0, uatom0}.Sort()
	globalFeesMix := sdk.Coins{photon0, uatom1}.Sort()

	tests := map[string]struct {
		globalfees          sdk.Coins
		zeroGlobalFeesDenom map[string]bool
		globalfeesNonZero   sdk.Coins
	}{
		"empty global fees": {
			globalfees:          globalFeesEmpty,
			zeroGlobalFeesDenom: map[string]bool{},
			globalfeesNonZero:   sdk.Coins{},
		},
		"nonzero coins global fees": {
			globalfees:          globalFees,
			zeroGlobalFeesDenom: map[string]bool{},
			globalfeesNonZero:   globalFees,
		},
		"zero coins global fees": {
			globalfees: globalFeesZeroCoins,
			zeroGlobalFeesDenom: map[string]bool{
				"photon": true,
				"uatom":  true,
			},
			globalfeesNonZero: sdk.Coins{},
		},
		"mix zero, nonzero coins global fees": {
			globalfees: globalFeesMix,
			zeroGlobalFeesDenom: map[string]bool{
				"photon": true,
			},
			globalfeesNonZero: sdk.NewCoins(uatom1),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			nonZeroCoins, zeroCoinsMap := getNonZeroFees(test.globalfees)
			require.True(t, nonZeroCoins.Equal(test.globalfeesNonZero))
			require.True(t, equalMap(zeroCoinsMap, test.zeroGlobalFeesDenom))
		})
	}
}

func equalMap(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) == 0 {
		return false
	}

	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}

	return true
}
