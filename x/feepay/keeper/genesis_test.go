package keeper_test

import (
	"fmt"

	"github.com/CosmosContracts/juno/v27/x/feepay/types"
)

func (s *KeeperTestSuite) TestFeeShareInitGenesis() {
	testCases := []struct {
		name    string
		genesis types.GenesisState
	}{
		{
			"Default Genesis - FeePay Enabled",
			s.genesis,
		},
		{
			"Custom Genesis - FeePay Disabled",
			types.GenesisState{
				Params: types.Params{
					EnableFeepay: false,
				},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.name), func() {
			s.SetupTest() // reset

			s.Require().NotPanics(func() {
				s.App.AppKeepers.FeePayKeeper.InitGenesis(s.Ctx, tc.genesis)
			})

			params := s.App.AppKeepers.FeePayKeeper.GetParams(s.Ctx)
			s.Require().Equal(tc.genesis.Params, params)
		})
	}
}
