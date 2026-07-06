package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v30/x/feemarket/types"
)

func TestMsgParams(t *testing.T) {
	t.Run("should reject a message with an invalid authority address", func(t *testing.T) {
		msg := types.NewMsgParams("invalid", types.DefaultParams())
		err := msg.ValidateBasic()
		require.Error(t, err)
	})

	t.Run("should accept an empty message with a valid authority address", func(t *testing.T) {
		msg := types.NewMsgParams(sdk.AccAddress("test").String(), types.DefaultParams())
		err := msg.ValidateBasic()
		require.NoError(t, err)
	})

	authority := sdk.AccAddress("test").String()

	invalidParamsCases := []struct {
		name     string
		malleate func(p *types.Params)
	}{
		{
			name:     "zero window",
			malleate: func(p *types.Params) { p.Window = 0 },
		},
		{
			name:     "empty fee denom",
			malleate: func(p *types.Params) { p.FeeDenom = "" },
		},
		{
			name:     "nil alpha",
			malleate: func(p *types.Params) { p.Alpha = math.LegacyDec{} },
		},
		{
			name:     "nil min base gas price",
			malleate: func(p *types.Params) { p.MinBaseGasPrice = math.LegacyDec{} },
		},
		{
			name:     "zero min base gas price",
			malleate: func(p *types.Params) { p.MinBaseGasPrice = math.LegacyZeroDec() },
		},
		{
			name:     "nil learning rates",
			malleate: func(p *types.Params) { p.MinLearningRate, p.MaxLearningRate = math.LegacyDec{}, math.LegacyDec{} },
		},
		{
			name: "min learning rate greater than max learning rate",
			malleate: func(p *types.Params) {
				p.MinLearningRate = math.LegacyMustNewDecFromStr("0.5")
				p.MaxLearningRate = math.LegacyMustNewDecFromStr("0.1")
			},
		},
		{
			name:     "invalid max block utilization",
			malleate: func(p *types.Params) { p.MaxBlockUtilization = 1 },
		},
		{
			name:     "negative beta",
			malleate: func(p *types.Params) { p.Beta = math.LegacyMustNewDecFromStr("-0.1") },
		},
	}

	for _, tc := range invalidParamsCases {
		t.Run("should reject invalid params: "+tc.name, func(t *testing.T) {
			params := types.DefaultParams()
			tc.malleate(&params)

			msg := types.NewMsgParams(authority, params)
			require.Error(t, msg.ValidateBasic())
		})
	}
}
