package keeper_test

import (
	"github.com/CosmosContracts/juno/v30/x/feemarket/types"
)

func (s *KeeperTestSuite) TestMsgParams() {
	s.Run("accepts a req with no params", func() {
		req := &types.MsgParams{
			Authority: s.authorityAccount.String(),
		}
		resp, err := s.msgServer.Params(s.Ctx, req)
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		params, err := s.App.AppKeepers.FeeMarketKeeper.GetParams(s.Ctx)
		s.Require().NoError(err)
		s.Require().False(params.Enabled)
	})

	s.Run("accepts a req with params", func() {
		req := &types.MsgParams{
			Authority: s.authorityAccount.String(),
			Params:    types.DefaultParams(),
		}
		resp, err := s.msgServer.Params(s.Ctx, req)
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		params, err := s.App.AppKeepers.FeeMarketKeeper.GetParams(s.Ctx)
		s.Require().NoError(err)
		s.Require().Equal(req.Params, params)
	})

	s.Run("rejects a req with invalid signer", func() {
		req := &types.MsgParams{
			Authority: "invalid",
		}
		_, err := s.msgServer.Params(s.Ctx, req)
		s.Require().Error(err)
	})

	s.Run("sets enabledHeight when transitioning from disabled -> enabled", func() {
		s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight())
		enabledParams := types.DefaultParams()

		req := &types.MsgParams{
			Authority: s.authorityAccount.String(),
			Params:    enabledParams,
		}
		_, err := s.msgServer.Params(s.Ctx, req)
		s.Require().NoError(err)

		disableParams := types.DefaultParams()
		disableParams.Enabled = false

		req = &types.MsgParams{
			Authority: s.authorityAccount.String(),
			Params:    disableParams,
		}
		_, err = s.msgServer.Params(s.Ctx, req)
		s.Require().NoError(err)

		gotHeight, err := s.App.AppKeepers.FeeMarketKeeper.GetEnabledHeight(s.Ctx)
		s.Require().NoError(err)
		s.Require().Equal(s.Ctx.BlockHeight(), gotHeight)

		// now that the markets are disabled, enable and check block height
		s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 10)

		req = &types.MsgParams{
			Authority: s.authorityAccount.String(),
			Params:    enabledParams,
		}
		_, err = s.msgServer.Params(s.Ctx, req)
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
		req := &types.MsgParams{
			Authority: s.authorityAccount.String(),
			Params:    params,
		}
		_, err = s.msgServer.Params(s.Ctx, req)
		s.Require().NoError(err)

		state, err = s.App.AppKeepers.FeeMarketKeeper.GetState(s.Ctx)
		s.Require().NoError(err)
		s.Require().Equal(params.Window, uint64(len(state.Window)))
		s.Require().Equal(state.Window[0], uint64(0))
	})
}
