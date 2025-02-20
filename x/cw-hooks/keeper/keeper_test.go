package keeper_test

import (
	"embed"
	"testing"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/stretchr/testify/suite"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/CosmosContracts/juno/v28/testutil"
	"github.com/CosmosContracts/juno/v28/x/cw-hooks/keeper"
	"github.com/CosmosContracts/juno/v28/x/cw-hooks/types"
)

var _ = embed.FS{}

//go:embed testdata/juno_staking_hooks_example.wasm
var wasmContract []byte

type KeeperTestSuite struct {
	testutil.KeeperTestHelper

	bankKeeper    bankkeeper.Keeper
	stakingKeeper stakingkeeper.Keeper
	wasmKeeper    wasmkeeper.Keeper

	queryClient   types.QueryClient
	msgServer     types.MsgServer
	wasmMsgServer wasmtypes.MsgServer
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()

	s.bankKeeper = s.App.AppKeepers.BankKeeper
	s.stakingKeeper = *s.App.AppKeepers.StakingKeeper
	s.wasmKeeper = s.App.AppKeepers.WasmKeeper

	s.queryClient = types.NewQueryClient(s.QueryHelper)
	s.msgServer = keeper.NewMsgServerImpl(s.App.AppKeepers.CWHooksKeeper)
	s.wasmMsgServer = wasmkeeper.NewMsgServerImpl(&s.wasmKeeper)
}
