package clock_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v18/app"
	clock "github.com/CosmosContracts/juno/v18/x/clock"
	"github.com/CosmosContracts/juno/v18/x/clock/types"
)

type GenesisTestSuite struct {
	suite.Suite

	ctx sdk.Context

	app *app.App
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
}

func (suite *GenesisTestSuite) TestClockInitGenesis() {
	_, _, addr := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()

	defaultParams := types.DefaultParams()

	testCases := []struct {
		name     string
		genesis  types.GenesisState
		expPanic bool
	}{
		{
			"default genesis",
			*clock.DefaultGenesisState(),
			false,
		},
		{
			"custom genesis - none",
			types.GenesisState{
				Params: types.Params{
					ContractGasLimit: defaultParams.ContractGasLimit,
				},
				ContractAddresses:       []string(nil),
				JailedContractAddresses: []string(nil),
			},
			false,
		},
		{
			"custom genesis - incorrect addr",
			types.GenesisState{
				Params: types.Params{
					ContractGasLimit: defaultParams.ContractGasLimit,
				},
				ContractAddresses:       []string{"incorrectaddr"},
				JailedContractAddresses: []string(nil),
			},
			true,
		},
		{
			"custom genesis - only one addr allowed",
			types.GenesisState{
				Params: types.Params{
					ContractGasLimit: defaultParams.ContractGasLimit,
				},
				ContractAddresses:       []string{addr.String(), addr2.String()},
				JailedContractAddresses: []string(nil),
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			if tc.expPanic {
				suite.Require().Panics(func() {
					clock.InitGenesis(suite.ctx, suite.app.AppKeepers.ClockKeeper, tc.genesis)
				})
			} else {
				suite.Require().NotPanics(func() {
					clock.InitGenesis(suite.ctx, suite.app.AppKeepers.ClockKeeper, tc.genesis)
				})

				params := suite.app.AppKeepers.ClockKeeper.GetParams(suite.ctx)
				suite.Require().Equal(tc.genesis.Params, params)
			}
		})
	}
}
