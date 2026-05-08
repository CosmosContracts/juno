package keeper_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	abci "github.com/cometbft/cometbft/abci/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/CosmosContracts/juno/v30/testutil"
	"github.com/CosmosContracts/juno/v30/x/stream/types"
	"github.com/CosmosContracts/juno/v30/x/stream/types/encoding"
)

type KeeperTestSuite struct {
	testutil.KeeperTestHelper

	bankKeeper    bankkeeper.Keeper
	stakingKeeper *stakingkeeper.Keeper

	queryClient types.QueryClient
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()

	s.bankKeeper = s.App.AppKeepers.BankKeeper
	s.stakingKeeper = s.App.AppKeepers.StakingKeeper

	s.queryClient = types.NewQueryClient(s.QueryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// TestSubscriptionRegistry tests the subscription registry functionality
func (s *KeeperTestSuite) TestSubscriptionRegistry() {
	ctx := context.Background()

	// Create a registry with a logger
	logger := s.App.Logger()
	registry := types.NewSubscriptionRegistry(logger, types.StreamConfig{})

	// Create a subscription
	subKey := encoding.StreamEvent{
		Module: "bank",
		Method: "balance",
		Params: map[string]string{
			"address": s.TestAccs[0].String(),
			"denom":   "ujuno",
		},
	}
	sendCh, subscriber := registry.Subscribe(ctx, subKey)
	s.Require().NotNil(subscriber)

	// Create an event and fan out
	event := encoding.StreamEvent{
		Module: "bank",
		Method: "balance",
		Params: map[string]string{
			"address": s.TestAccs[0].String(),
			"denom":   "ujuno",
		},
	}

	// Fan out the event with the event data itself
	registry.FanOut(event, event)

	// Check if event was received
	select {
	case received := <-sendCh:
		receivedEvent, ok := received.(encoding.StreamEvent)
		s.Require().True(ok)
		addr, ok := receivedEvent.Params["address"]
		s.Require().True(ok)
		s.Require().Equal(s.TestAccs[0].String(), addr)
		denom, ok := receivedEvent.Params["denom"]
		s.Require().True(ok)
		s.Require().Equal("ujuno", denom)
	case <-time.After(100 * time.Millisecond):
		s.Fail("Did not receive event in time")
	}

	// Unsubscribe
	registry.Unsubscribe(subscriber)
}

// TestStreamingListener tests the ABCI listener functionality
func (s *KeeperTestSuite) TestStreamingListener() {
	intake := s.App.AppKeepers.StreamKeeper.Dispatcher().Intake()
	listener := types.NewStreamingListener(
		intake,
		s.App.Logger(),
		// nil,
	)
	s.Require().NotNil(listener)

	// Test ListenFinalizeBlock (should not error)
	err := listener.ListenFinalizeBlock(context.Background(), abci.RequestFinalizeBlock{}, abci.ResponseFinalizeBlock{})
	s.Require().NoError(err)

	// Test ListenCommit (should not error)
	err = listener.ListenCommit(s.Ctx, abci.ResponseCommit{}, nil)
	s.Require().NoError(err)
}

// TestIntakeChannel tests the intake channel functionality
func (s *KeeperTestSuite) TestIntakeChannel() {
	intake := s.App.AppKeepers.StreamKeeper.Dispatcher().Intake()
	s.Require().NotNil(intake)

	// Test sending an event to the intake channel
	event := encoding.StreamEvent{
		Module: "bank",
		Method: "balance",
		Params: map[string]string{
			"address": s.TestAccs[0].String(),
			"denom":   "ujuno",
		},
	}

	// Send event (should not block)
	select {
	case intake <- event:
	case <-time.After(100 * time.Millisecond):
		s.Fail("Failed to send event to intake channel")
	}
}

func (s *KeeperTestSuite) TestExecuteQueryBalance() {
	keeper := s.App.AppKeepers.StreamKeeper
	s.Commit()
	registry := keeper.MethodRegistry()
	s.Require().NotNil(registry)

	desc, ok := registry.LookupByModuleAndStream("bank", "balance")
	s.Require().True(ok)
	s.Require().NotNil(desc)

	addr := s.TestAccs[0]
	fields := map[string]string{
		"address": addr.String(),
		"denom":   "ujuno",
	}

	resp, err := keeper.Invoker().Execute(desc, fields)
	s.Require().NoError(err)

	balanceResp, ok := resp.(*banktypes.QueryBalanceResponse)
	s.Require().NotNil(resp)
	s.Require().True(ok, "unexpected response type: %T", resp)
	s.Require().Equal("ujuno", balanceResp.Balance.Denom)
}
