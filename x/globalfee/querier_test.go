package globalfee

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	globalfeekeeper "github.com/CosmosContracts/juno/v16/x/globalfee/keeper"
	"github.com/CosmosContracts/juno/v16/x/globalfee/types"
)

func TestQueryMinimumGasPrices(t *testing.T) {
	specs := map[string]struct {
		setupStore func(ctx sdk.Context, k globalfeekeeper.Keeper)
		expMin     sdk.DecCoins
	}{
		"one coin": {
			setupStore: func(ctx sdk.Context, k globalfeekeeper.Keeper) {
				k.SetParams(ctx, types.Params{
					MinimumGasPrices: sdk.NewDecCoins(sdk.NewDecCoin("ALX", sdk.OneInt())),
				})
			},
			expMin: sdk.NewDecCoins(sdk.NewDecCoin("ALX", sdk.OneInt())),
		},
		"multiple coins": {
			setupStore: func(ctx sdk.Context, k globalfeekeeper.Keeper) {
				k.SetParams(ctx, types.Params{
					MinimumGasPrices: sdk.NewDecCoins(sdk.NewDecCoin("ALX", sdk.OneInt()), sdk.NewDecCoin("BLX", sdk.NewInt(2))),
				})
			},
			expMin: sdk.NewDecCoins(sdk.NewDecCoin("ALX", sdk.OneInt()), sdk.NewDecCoin("BLX", sdk.NewInt(2))),
		},
		"no min gas price set": {
			setupStore: func(ctx sdk.Context, k globalfeekeeper.Keeper) {
				k.SetParams(ctx, types.Params{})
			},
		},
		"no param set": {
			setupStore: func(ctx sdk.Context, k globalfeekeeper.Keeper) {
			},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, _, keeper := setupTestStore(t)
			spec.setupStore(ctx, keeper)
			q := NewGrpcQuerier(keeper)
			gotResp, gotErr := q.MinimumGasPrices(sdk.WrapSDKContext(ctx), nil)
			require.NoError(t, gotErr)
			require.NotNil(t, gotResp)
			assert.Equal(t, spec.expMin, gotResp.MinimumGasPrices)
		})
	}
}
