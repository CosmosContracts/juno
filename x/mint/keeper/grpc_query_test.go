package keeper_test

import (
	gocontext "context"
	"testing"

	"github.com/stretchr/testify/suite"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v16/app"
	"github.com/CosmosContracts/juno/v16/x/mint/types"
)

type MintTestSuite struct {
	suite.Suite

	app         *app.App
	ctx         sdk.Context
	queryClient types.QueryClient
}

func (suite *MintTestSuite) SetupTest() {
	app := app.Setup(suite.T())
	ctx := app.BaseApp.NewContext(false, tmproto.Header{
		ChainID: "testing",
	})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.AppKeepers.MintKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.app = app
	suite.ctx = ctx

	suite.queryClient = queryClient
}

func (suite *MintTestSuite) TestGRPCParams() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient

	params, err := queryClient.Params(gocontext.Background(), &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(params.Params, app.AppKeepers.MintKeeper.GetParams(ctx))

	inflation, err := queryClient.Inflation(gocontext.Background(), &types.QueryInflationRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(inflation.Inflation, app.AppKeepers.MintKeeper.GetMinter(ctx).Inflation)

	annualProvisions, err := queryClient.AnnualProvisions(gocontext.Background(), &types.QueryAnnualProvisionsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(annualProvisions.AnnualProvisions, app.AppKeepers.MintKeeper.GetMinter(ctx).AnnualProvisions)
}

func TestMintTestSuite(t *testing.T) {
	suite.Run(t, new(MintTestSuite))
}
