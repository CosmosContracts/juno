package feepay_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v19/app"
	"github.com/CosmosContracts/juno/v19/x/feepay"
	"github.com/CosmosContracts/juno/v19/x/feepay/types"
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

func (suite *GenesisTestSuite) TestFeeShareInitGenesis() {
	testCases := []struct {
		name    string
		genesis types.GenesisState
	}{
		{
			"Default Genesis - FeePay Enabled",
			suite.genesis,
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
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			suite.Require().NotPanics(func() {
				feepay.InitGenesis(suite.ctx, suite.app.AppKeepers.FeePayKeeper, tc.genesis)
			})

			params := suite.app.AppKeepers.FeePayKeeper.GetParams(suite.ctx)
			suite.Require().Equal(tc.genesis.Params, params)
		})
	}
}
