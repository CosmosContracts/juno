package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	_ "embed"

	"github.com/CosmosContracts/juno/v30/testutil"
	"github.com/CosmosContracts/juno/v30/x/clock/keeper"
	"github.com/CosmosContracts/juno/v30/x/clock/types"
)

type KeeperTestSuite struct {
	testutil.KeeperTestHelper

	queryClient types.QueryClient
	msgServer   types.MsgServer
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
	s.queryClient = types.NewQueryClient(s.QueryHelper)
	s.msgServer = keeper.NewMsgServerImpl(s.App.AppKeepers.ClockKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

//go:embed testdata/clock_example.wasm
var clockContract []byte

//go:embed testdata/cw_testburn.wasm
var burnContract []byte

// Helper method for quickly registering a clock contract
func (s *KeeperTestSuite) RegisterClockContract(senderAddress string, contractAddress string) {
	err := s.App.AppKeepers.ClockKeeper.RegisterContract(s.Ctx, senderAddress, contractAddress)
	s.Require().NoError(err)
}

// Helper method for quickly unregistering a clock contract
func (s *KeeperTestSuite) UnregisterClockContract(senderAddress string, contractAddress string) {
	err := s.App.AppKeepers.ClockKeeper.UnregisterContract(s.Ctx, senderAddress, contractAddress)
	s.Require().NoError(err)
}

// Helper method for quickly jailing a clock contract
func (s *KeeperTestSuite) JailClockContract(contractAddress string) {
	err := s.App.AppKeepers.ClockKeeper.SetJailStatus(s.Ctx, contractAddress, true)
	s.Require().NoError(err)
}

// Helper method for quickly unjailing a clock contract
func (s *KeeperTestSuite) UnjailClockContract(senderAddress string, contractAddress string) {
	err := s.App.AppKeepers.ClockKeeper.SetJailStatusBySender(s.Ctx, senderAddress, contractAddress, false)
	s.Require().NoError(err)
}
