package keeper_test

import (
	"testing"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/stretchr/testify/suite"

	_ "embed"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/CosmosContracts/juno/v30/testutil"
	"github.com/CosmosContracts/juno/v30/x/feeshare/types"
)

type KeeperTestSuite struct {
	testutil.KeeperTestHelper

	genesis types.GenesisState

	bankKeeper    bankkeeper.Keeper
	accountKeeper authkeeper.AccountKeeper

	queryClient   types.QueryClient
	msgServer     types.MsgServer
	wasmMsgServer wasmtypes.MsgServer
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
	s.genesis = *types.DefaultGenesisState()

	s.bankKeeper = s.App.AppKeepers.BankKeeper
	s.accountKeeper = s.App.AppKeepers.AccountKeeper

	s.queryClient = types.NewQueryClient(s.QueryHelper)
	s.msgServer = s.App.AppKeepers.FeeShareKeeper
	s.wasmMsgServer = wasmkeeper.NewMsgServerImpl(&s.App.AppKeepers.WasmKeeper)
}

//go:embed testdata/reflect.wasm
var wasmContract []byte
