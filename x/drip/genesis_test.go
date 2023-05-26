package drip_test

import (
	"fmt"
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/suite"

	"github.com/CosmosContracts/juno/v15/app"
	drip "github.com/CosmosContracts/juno/v15/x/drip"
	"github.com/CosmosContracts/juno/v15/x/drip/types"
)

type GenesisTestSuite struct {
	suite.Suite

	ctx sdk.Context

	app     *app.App
	genesis types.GenesisState
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) SetupTest() {
	app := app.Setup(suite.T())
	ctx := app.BaseApp.NewContext(false, tmproto.Header{
		ChainID: "testing",
	})

	suite.app = app
	suite.ctx = ctx

	suite.genesis = *types.DefaultGenesisState()
}

func (suite *GenesisTestSuite) TestDripInitGenesis() {
	testCases := []struct {
		name     string
		genesis  types.GenesisState
		expPanic bool
	}{
		{
			"default genesis",
			suite.genesis,
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
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			if tc.expPanic {
				suite.Require().Panics(func() {
					drip.InitGenesis(suite.ctx, suite.app.AppKeepers.DripKeeper, tc.genesis)
				})
			} else {
				suite.Require().NotPanics(func() {
					drip.InitGenesis(suite.ctx, suite.app.AppKeepers.DripKeeper, tc.genesis)
				})

				params := suite.app.AppKeepers.DripKeeper.GetParams(suite.ctx)
				suite.Require().Equal(tc.genesis.Params, params)
			}
		})
	}
}
