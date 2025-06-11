package keeper_test

import (
	"fmt"

	"github.com/CosmosContracts/juno/v30/x/clock/types"
)

func (s *KeeperTestSuite) TestClockInitGenesis() {
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
			s.Reset()
			if tc.success {
				s.Require().NotPanics(func() {
					s.App.AppKeepers.ClockKeeper.InitGenesis(s.Ctx, tc.genesis)
				})

				params := s.App.AppKeepers.ClockKeeper.GetParams(s.Ctx)
				s.Require().Equal(tc.genesis.Params, params)
			} else {
				s.Require().Panics(func() {
					s.App.AppKeepers.ClockKeeper.InitGenesis(s.Ctx, tc.genesis)
				})
			}
		})
	}
}
