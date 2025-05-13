package keeper_test

import (
	"testing"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/stretchr/testify/suite"

	_ "embed"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/CosmosContracts/juno/v29/testutil"
	"github.com/CosmosContracts/juno/v29/x/feepay/keeper"
	"github.com/CosmosContracts/juno/v29/x/feepay/types"
)

type KeeperTestSuite struct {
	testutil.KeeperTestHelper

	genesis types.GenesisState

	bankKeeper bankkeeper.Keeper

	queryClient   types.QueryClient
	msgServer     types.MsgServer
	wasmMsgServer wasmtypes.MsgServer
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
	s.genesis = *types.DefaultGenesisState()

	s.bankKeeper = s.App.AppKeepers.BankKeeper

	s.queryClient = types.NewQueryClient(s.QueryHelper)
	s.msgServer = keeper.NewMsgServerImpl(s.App.AppKeepers.FeePayKeeper)
	s.wasmMsgServer = wasmkeeper.NewMsgServerImpl(&s.App.AppKeepers.WasmKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

//go:embed testdata/clock_example.wasm
var wasmContract []byte

// Helper method for quickly registering a fee pay contract
func (s *KeeperTestSuite) registerFeePayContract(senderAddress string, contractAddress string, balance uint64, walletLimit uint64) {
	_, err := s.msgServer.RegisterFeePayContract(s.Ctx, &types.MsgRegisterFeePayContract{
		SenderAddress: senderAddress,
		FeePayContract: &types.FeePayContract{
			ContractAddress: contractAddress,
			Balance:         balance,
			WalletLimit:     walletLimit,
		},
	})
	s.Require().NoError(err)
}
