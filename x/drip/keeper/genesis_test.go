package keeper_test

import (
	"fmt"

	"github.com/CosmosContracts/juno/v29/x/drip/types"
)

func (s *KeeperTestSuite) TestDripInitGenesis() {
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
			"custom genesis - drip enabled, no one allowed",
			types.GenesisState{
				Params: types.Params{
					EnableDrip:       true,
					AllowedAddresses: []string(nil),
				},
			},
			false,
		},
		{
			"custom genesis - drip enabled, only one addr allowed",
			types.GenesisState{
				Params: types.Params{
					EnableDrip:       true,
					AllowedAddresses: []string{"juno1v6vlpuqlhhpwujvaqs4pe5dmljapdev4s827ql"},
				},
			},
			false,
		},
		{
			"custom genesis - drip enabled, 2 addr allowed",
			types.GenesisState{
				Params: types.Params{
					EnableDrip:       true,
					AllowedAddresses: []string{"juno1v6vlpuqlhhpwujvaqs4pe5dmljapdev4s827ql", "juno1hq2p69p4kmwndxlss7dqk0sr5pe5mmcpf7wqec"},
				},
			},
			false,
		},
		{
			"custom genesis - drip enabled, address invalid",
			types.GenesisState{
				Params: types.Params{
					EnableDrip:       true,
					AllowedAddresses: []string{"juno1v6vllollollollollolloldmljapdev4s827ql"},
				},
			},
			true,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.name), func() {
			s.SetupTest() // reset

			if tc.expPanic {
				s.Require().Panics(func() {
					s.App.AppKeepers.DripKeeper.InitGenesis(s.Ctx, tc.genesis)
				})
			} else {
				s.Require().NotPanics(func() {
					s.App.AppKeepers.DripKeeper.InitGenesis(s.Ctx, tc.genesis)
				})

				params := s.App.AppKeepers.DripKeeper.GetParams(s.Ctx)
				s.Require().Equal(tc.genesis.Params, params)
			}
		})
	}
}
