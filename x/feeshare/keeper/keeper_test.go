package keeper_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/CosmosContracts/juno/v12/app"
	"github.com/CosmosContracts/juno/v12/x/feeshare/keeper"
	"github.com/CosmosContracts/juno/v12/x/feeshare/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	ctx               sdk.Context
	app               *app.App
	queryClient       types.QueryClient
	feeShareMsgServer types.MsgServer
	wasmMsgServer     wasmtypes.MsgServer
}

func (s *IntegrationTestSuite) SetupTest() {
	isCheckTx := false
	s.app = app.Setup(isCheckTx)

	s.ctx = s.app.BaseApp.NewContext(isCheckTx, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
		Height:  9,
	})

	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQuerier(s.app.FeeShareKeeper))

	s.queryClient = types.NewQueryClient(queryHelper)
	s.feeShareMsgServer = s.app.FeeShareKeeper
	s.wasmMsgServer = wasmkeeper.NewMsgServerImpl(wasmkeeper.NewDefaultPermissionKeeper(s.app.WasmKeeper))
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
