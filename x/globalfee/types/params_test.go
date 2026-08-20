package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestDecCoinsValidatePreservesErrorMessages(t *testing.T) {
	tests := []struct {
		name  string
		coins DecCoins
		want  string
	}{
		{
			name: "duplicate denomination",
			coins: DecCoins{
				sdk.NewDecCoinFromDec("aaa", math.LegacyOneDec()),
				sdk.NewDecCoinFromDec("aaa", math.LegacyOneDec()),
			},
			want: "duplicate denomination aaa",
		},
		{
			name: "unsorted denomination",
			coins: DecCoins{
				sdk.NewDecCoinFromDec("bbb", math.LegacyOneDec()),
				sdk.NewDecCoinFromDec("aaa", math.LegacyOneDec()),
			},
			want: "denomination aaa is not sorted",
		},
		{
			name: "negative amount",
			coins: DecCoins{
				{Denom: "aaa", Amount: math.LegacyNewDec(-1)},
			},
			want: "coin -1.000000000000000000 amount is negative",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.EqualError(t, tc.coins.Validate(), tc.want)
		})
	}
}
