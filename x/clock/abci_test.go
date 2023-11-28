package clock_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v18/app"
	clock "github.com/CosmosContracts/juno/v18/x/clock"
)

type EndBlockerTestSuite struct {
	suite.Suite

	ctx sdk.Context

	app *app.App
}

func TestEndBlockerTestSuite(t *testing.T) {
	suite.Run(t, new(EndBlockerTestSuite))
}

func (s *EndBlockerTestSuite) SetupTest() {
	app := app.Setup(s.T())
	ctx := app.BaseApp.NewContext(false, tmproto.Header{
		ChainID: "testing",
	})

	s.app = app
	s.ctx = ctx
}

func (s *EndBlockerTestSuite) TestEndBlocker() {
	// Call end blocker
	clock.EndBlocker(s.ctx, s.app.AppKeepers.ClockKeeper)
}
