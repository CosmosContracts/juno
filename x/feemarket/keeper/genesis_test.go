package keeper_test

import (
	"github.com/CosmosContracts/juno/v30/x/feemarket/types"
)

func (s *KeeperTestSuite) TestInitGenesis() {
	s.Run("default genesis should not panic", func() {
		s.Require().NotPanics(func() {
			s.App.AppKeepers.FeeMarketKeeper.InitGenesis(s.Ctx, *types.DefaultGenesisState())
		})
	})

	s.Run("default AIMD genesis should not panic", func() {
		s.Require().NotPanics(func() {
			s.App.AppKeepers.FeeMarketKeeper.InitGenesis(s.Ctx, *types.DefaultAIMDGenesisState())
		})
	})

	s.Run("bad genesis state should panic", func() {
		gs := types.DefaultGenesisState()
		gs.Params.Window = 0
		s.Require().Panics(func() {
			s.App.AppKeepers.FeeMarketKeeper.InitGenesis(s.Ctx, *gs)
		})
	})

	s.Run("mismatch in params and state for window should panic", func() {
		gs := types.DefaultAIMDGenesisState()
		gs.Params.Window = 1

		s.Require().Panics(func() {
			s.App.AppKeepers.FeeMarketKeeper.InitGenesis(s.Ctx, *gs)
		})
	})
}

func (s *KeeperTestSuite) TestExportGenesis() {
	s.Run("export genesis should not panic for default eip-1559", func() {
		gs := types.DefaultGenesisState()
		s.App.AppKeepers.FeeMarketKeeper.InitGenesis(s.Ctx, *gs)

		var exportedGenesis *types.GenesisState
		s.Require().NotPanics(func() {
			exportedGenesis = s.App.AppKeepers.FeeMarketKeeper.ExportGenesis(s.Ctx)
		})

		s.Require().Equal(gs, exportedGenesis)
	})

	s.Run("export genesis should not panic for default AIMD eip-1559", func() {
		gs := types.DefaultAIMDGenesisState()
		s.App.AppKeepers.FeeMarketKeeper.InitGenesis(s.Ctx, *gs)

		var exportedGenesis *types.GenesisState
		s.Require().NotPanics(func() {
			exportedGenesis = s.App.AppKeepers.FeeMarketKeeper.ExportGenesis(s.Ctx)
		})

		s.Require().Equal(gs, exportedGenesis)
	})
}
