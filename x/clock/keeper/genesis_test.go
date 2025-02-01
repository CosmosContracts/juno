package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v27/app"
	"github.com/CosmosContracts/juno/v27/testutil"
	"github.com/CosmosContracts/juno/v27/x/clock/types"
)

type GenesisTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *app.App
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (s *GenesisTestSuite) SetupTest() {
	app := testutil.Setup(false, s.T())
	ctx := app.BaseApp.NewContext(false)

	s.app = app
	s.ctx = ctx
}

func (s *GenesisTestSuite) TestClockInitGenesis() {
	testCases := []struct {
		name    string
		genesis types.GenesisState
		success bool
	}{
		{
			"Success - Default Genesis",
			*types.DefaultGenesisState(),
			true,
		},
		{
			"Success - Custom Genesis",
			types.GenesisState{
				Params: types.Params{
					ContractGasLimit: 500_000,
				},
			},
			true,
		},
		{
			"Fail - Invalid Gas Amount",
			types.GenesisState{
				Params: types.Params{
					ContractGasLimit: 1,
				},
			},
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.name), func() {
			s.SetupTest() // reset

			if tc.success {
				s.Require().NotPanics(func() {
					s.app.AppKeepers.ClockKeeper.InitGenesis(s.ctx, tc.genesis)
				})

				params := s.app.AppKeepers.ClockKeeper.GetParams(s.ctx)
				s.Require().Equal(tc.genesis.Params, params)
			} else {
				s.Require().Panics(func() {
					s.app.AppKeepers.ClockKeeper.InitGenesis(s.ctx, tc.genesis)
				})
			}
		})
	}
}
