package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	"github.com/CosmosContracts/juno/v31/x/feemarket/types"
	feepaytypes "github.com/CosmosContracts/juno/v31/x/feepay/types"
)

func (s *KeeperTestSuite) TestMsgParams() {
	s.Run("rejects a req with no params", func() {
		// empty params (zero window, nil decimals, empty fee denom) would
		// panic the ante/post handlers and EndBlock if ever stored.
		req := &types.MsgUpdateParams{
			Authority: s.authorityAccount.String(),
		}
		_, err := s.msgServer.UpdateParams(s.Ctx, req)
		s.Require().Error(err)
	})

	s.Run("accepts a req with params", func() {
		req := &types.MsgUpdateParams{
			Authority: s.authorityAccount.String(),
			Params:    types.DefaultParams(),
		}
		resp, err := s.msgServer.UpdateParams(s.Ctx, req)
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		params, err := s.App.AppKeepers.FeeMarketKeeper.GetParams(s.Ctx)
		s.Require().NoError(err)
		s.Require().Equal(req.Params, params)
	})

	s.Run("rejects fee denom change while FeePay has outstanding balances", func() {
		before, err := s.App.AppKeepers.FeeMarketKeeper.GetParams(s.Ctx)
		s.Require().NoError(err)

		s.App.AppKeepers.FeePayKeeper.SetFeePayContract(s.Ctx, feepaytypes.FeePayContract{
			ContractAddress: s.authorityAccount.String(),
			Balance:         1,
		})

		changed := before
		changed.FeeDenom = "uother"
		_, err = s.msgServer.UpdateParams(s.Ctx, &types.MsgUpdateParams{
			Authority: s.authorityAccount.String(),
			Params:    changed,
		})
		s.Require().ErrorContains(err, "outstanding FeePay balances")

		after, err := s.App.AppKeepers.FeeMarketKeeper.GetParams(s.Ctx)
		s.Require().NoError(err)
		s.Require().Equal(before, after)

		s.App.AppKeepers.FeePayKeeper.SetFeePayContract(s.Ctx, feepaytypes.FeePayContract{
			ContractAddress: s.authorityAccount.String(),
			Balance:         0,
		})
		_, err = s.msgServer.UpdateParams(s.Ctx, &types.MsgUpdateParams{
			Authority: s.authorityAccount.String(),
			Params:    changed,
		})
		s.Require().NoError(err)
	})

	s.Run("rejects a req with invalid signer", func() {
		req := &types.MsgUpdateParams{
			Authority: "invalid",
		}
		_, err := s.msgServer.UpdateParams(s.Ctx, req)
		s.Require().Error(err)
	})

	s.Run("rejects invalid params and leaves state untouched", func() {
		invalidCases := []struct {
			name     string
			malleate func(p *types.Params)
		}{
			{"zero window", func(p *types.Params) { p.Window = 0 }},
			{"empty fee denom", func(p *types.Params) { p.FeeDenom = "" }},
			{"nil min base gas price", func(p *types.Params) { p.MinBaseGasPrice = sdkmath.LegacyDec{} }},
			{"zero min base gas price", func(p *types.Params) { p.MinBaseGasPrice = sdkmath.LegacyZeroDec() }},
			{"nil learning rates", func(p *types.Params) { p.MinLearningRate, p.MaxLearningRate = sdkmath.LegacyDec{}, sdkmath.LegacyDec{} }},
			{
				"min learning rate greater than max", func(p *types.Params) {
					p.MinLearningRate = sdkmath.LegacyMustNewDecFromStr("0.5")
					p.MaxLearningRate = sdkmath.LegacyMustNewDecFromStr("0.1")
				},
			},
			{"invalid max block utilization", func(p *types.Params) { p.MaxBlockUtilization = 1 }},
			{"nil alpha", func(p *types.Params) { p.Alpha = sdkmath.LegacyDec{} }},
		}

		for _, tc := range invalidCases {
			s.Run(tc.name, func() {
				gotParams, err := s.App.AppKeepers.FeeMarketKeeper.GetParams(s.Ctx)
				s.Require().NoError(err)
				gotState, err := s.App.AppKeepers.FeeMarketKeeper.GetState(s.Ctx)
				s.Require().NoError(err)

				invalid := types.DefaultParams()
				tc.malleate(&invalid)

				req := &types.MsgUpdateParams{
					Authority: s.authorityAccount.String(),
					Params:    invalid,
				}
				_, err = s.msgServer.UpdateParams(s.Ctx, req)
				s.Require().Error(err)

				// stored params and state must be unchanged
				params, err := s.App.AppKeepers.FeeMarketKeeper.GetParams(s.Ctx)
				s.Require().NoError(err)
				s.Require().Equal(gotParams, params)
				state, err := s.App.AppKeepers.FeeMarketKeeper.GetState(s.Ctx)
				s.Require().NoError(err)
				s.Require().Equal(gotState, state)
			})
		}
	})

	s.Run("sets enabledHeight when transitioning from disabled -> enabled", func() {
		// genesis params are enabled — disable first so the enable below is a
		// real disabled -> enabled transition
		disableParams := types.DefaultParams()
		disableParams.Enabled = false
		req := &types.MsgUpdateParams{
			Authority: s.authorityAccount.String(),
			Params:    disableParams,
		}
		_, err := s.msgServer.UpdateParams(s.Ctx, req)
		s.Require().NoError(err)

		enabledParams := types.DefaultParams()
		req = &types.MsgUpdateParams{
			Authority: s.authorityAccount.String(),
			Params:    enabledParams,
		}
		_, err = s.msgServer.UpdateParams(s.Ctx, req)
		s.Require().NoError(err)

		gotHeight, err := s.App.AppKeepers.FeeMarketKeeper.GetEnabledHeight(s.Ctx)
		s.Require().NoError(err)
		s.Require().Equal(s.Ctx.BlockHeight(), gotHeight)

		// disable again before testing the next transition
		req = &types.MsgUpdateParams{
			Authority: s.authorityAccount.String(),
			Params:    disableParams,
		}
		_, err = s.msgServer.UpdateParams(s.Ctx, req)
		s.Require().NoError(err)

		// now that the markets are disabled, enable and check block height
		s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 10)

		req = &types.MsgUpdateParams{
			Authority: s.authorityAccount.String(),
			Params:    enabledParams,
		}
		_, err = s.msgServer.UpdateParams(s.Ctx, req)
		s.Require().NoError(err)

		newHeight, err := s.App.AppKeepers.FeeMarketKeeper.GetEnabledHeight(s.Ctx)
		s.Require().NoError(err)
		s.Require().Equal(s.Ctx.BlockHeight(), newHeight)
	})

	s.Run("resets state after new params request", func() {
		params, err := s.App.AppKeepers.FeeMarketKeeper.GetParams(s.Ctx)
		s.Require().NoError(err)

		state, err := s.App.AppKeepers.FeeMarketKeeper.GetState(s.Ctx)
		s.Require().NoError(err)

		err = state.Update(params.MaxBlockUtilization, params)
		s.Require().NoError(err)

		err = s.App.AppKeepers.FeeMarketKeeper.SetState(s.Ctx, state)
		s.Require().NoError(err)

		params.Window = 100
		req := &types.MsgUpdateParams{
			Authority: s.authorityAccount.String(),
			Params:    params,
		}
		_, err = s.msgServer.UpdateParams(s.Ctx, req)
		s.Require().NoError(err)

		state, err = s.App.AppKeepers.FeeMarketKeeper.GetState(s.Ctx)
		s.Require().NoError(err)
		s.Require().Equal(params.Window, uint64(len(state.Window)))
		s.Require().Equal(state.Window[0], uint64(0))
	})
}
