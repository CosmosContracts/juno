package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/testutil"

	"github.com/CosmosContracts/juno/v27/x/mint/types"
)

func (s *KeeperTestSuite) TestImportExportGenesis() {
	genesisState := types.DefaultGenesisState()
	genesisState.Minter = types.InitialMinter(sdkmath.LegacyNewDecWithPrec(20, 2))
	genesisState.Params = types.NewParams(
		"testDenom",
		uint64(60*60*8766/5),
	)

	s.mintKeeper.InitGenesis(s.Ctx, s.accountKeeper, genesisState)

	minter, err := s.mintKeeper.GetMinter(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(genesisState.Minter, minter)

	invalidCtx := testutil.DefaultContextWithDB(s.T(), storetypes.NewKVStoreKey("mint"), storetypes.NewTransientStoreKey("transient_test"))
	_, err = s.mintKeeper.GetMinter(invalidCtx.Ctx)
	s.Require().Error(err)

	params, err := s.mintKeeper.GetParams(s.Ctx)
	s.Require().Equal(genesisState.Params, params)
	s.Require().NoError(err)

	genesisState2 := s.mintKeeper.ExportGenesis(s.Ctx)
	s.Require().Equal(genesisState, genesisState2)
}
