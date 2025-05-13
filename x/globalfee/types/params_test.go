package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v29/x/globalfee/types"
)

func TestDefaultParams(t *testing.T) {
	p := types.DefaultParams()
	require.EqualValues(t, p.MinimumGasPrices, sdk.DecCoins(nil))
}

func Test_validateParams(t *testing.T) {
	tests := map[string]struct {
		coins     interface{} // not sdk.DeCoins, but Decoins defined in glboalfee
		expectErr bool
	}{
		"DefaultParams, pass": {
			types.DefaultParams().MinimumGasPrices,
			false,
		},
		"DecCoins conversion fails, fail": {
			sdk.Coins{sdk.NewCoin("photon", sdkmath.OneInt())},
			true,
		},
		"coins amounts are zero, pass": {
			sdk.DecCoins{
				sdk.NewDecCoin("atom", sdkmath.ZeroInt()),
				sdk.NewDecCoin("photon", sdkmath.ZeroInt()),
			},
			false,
		},
		"duplicate coins denoms, fail": {
			sdk.DecCoins{
				sdk.NewDecCoin("photon", sdkmath.OneInt()),
				sdk.NewDecCoin("photon", sdkmath.OneInt()),
			},
			true,
		},
		"coins are not sorted by denom alphabetically, fail": {
			sdk.DecCoins{
				sdk.NewDecCoin("photon", sdkmath.OneInt()),
				sdk.NewDecCoin("atom", sdkmath.OneInt()),
			},
			true,
		},
		"negative amount, fail": {
			sdk.DecCoins{
				sdk.DecCoin{Denom: "photon", Amount: sdkmath.LegacyOneDec().Neg()},
			},
			true,
		},
		"invalid denom, fail": {
			sdk.DecCoins{
				sdk.DecCoin{Denom: "photon!", Amount: sdkmath.LegacyOneDec().Neg()},
			},
			true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := types.ValidateMinimumGasPrices(test.coins)
			if test.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
