package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	"github.com/CosmosContracts/juno/v27/x/feeshare/types"
)

func (s *KeeperTestSuite) TestFeeShareInitGenesis() {
	testCases := []struct {
		name     string
		genesis  types.GenesisState
		expPanic bool
	}{
		{
			"default genesis",
			s.genesis,
			false,
		},
		{
			"custom genesis - feeshare disabled",
			types.GenesisState{
				Params: types.Params{
					EnableFeeShare:  false,
					DeveloperShares: sdkmath.LegacyNewDecWithPrec(50, 2),
					AllowedDenoms:   []string{"ujuno"},
				},
			},
			false,
		},
		{
			"custom genesis - feeshare enabled, 0% developer shares",
			types.GenesisState{
				Params: types.Params{
					EnableFeeShare:  true,
					DeveloperShares: sdkmath.LegacyNewDecWithPrec(0, 2),
					AllowedDenoms:   []string{"ujuno"},
				},
			},
			false,
		},
		{
			"custom genesis - feeshare enabled, 100% developer shares",
			types.GenesisState{
				Params: types.Params{
					EnableFeeShare:  true,
					DeveloperShares: sdkmath.LegacyNewDecWithPrec(100, 2),
					AllowedDenoms:   []string{"ujuno"},
				},
			},
			false,
		},
		{
			"custom genesis - feeshare enabled, all denoms allowed",
			types.GenesisState{
				Params: types.Params{
					EnableFeeShare:  true,
					DeveloperShares: sdkmath.LegacyNewDecWithPrec(10, 2),
					AllowedDenoms:   []string(nil),
				},
			},
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.name), func() {
			s.SetupTest() // reset

			if tc.expPanic {
				s.Require().Panics(func() {
					s.app.AppKeepers.FeeShareKeeper.InitGenesis(s.ctx, tc.genesis)
				})
			} else {
				s.Require().NotPanics(func() {
					s.app.AppKeepers.FeeShareKeeper.InitGenesis(s.ctx, tc.genesis)
				})

				params := s.app.AppKeepers.FeeShareKeeper.GetParams(s.ctx)
				s.Require().Equal(tc.genesis.Params, params)
			}
		})
	}
}
