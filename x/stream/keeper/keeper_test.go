package keeper_test

import (
	"context"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/suite"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/CosmosContracts/juno/v30/testutil"
	"github.com/CosmosContracts/juno/v30/x/stream/keeper"
	"github.com/CosmosContracts/juno/v30/x/stream/types"
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
	registry := keeper.NewSubscriptionRegistry(logger)

	// Create a subscription
	subKey := types.GenerateSubscriptionKey(types.SubscriptionTypeBalance, s.TestAccs[0].String(), "", "ujuno")
	sendCh := make(chan any, 32)
	subscriber := registry.Subscribe(subKey, ctx, sendCh)
	s.Require().NotNil(subscriber)

	// Create an event and fan out
	event := types.StreamEvent{
		Module:      types.ModuleNameBank,
		EventType:   types.EventTypeBalanceChange,
		Address:     s.TestAccs[0].String(),
		Denom:       "ujuno",
		BlockHeight: 100,
	}

	// Fan out the event with the event data itself
	registry.FanOut(event, event)

	// Check if event was received
	select {
	case received := <-sendCh:
		receivedEvent, ok := received.(types.StreamEvent)
		s.Require().True(ok)
		s.Require().Equal(event.Address, receivedEvent.Address)
		s.Require().Equal(event.Denom, receivedEvent.Denom)
	case <-time.After(100 * time.Millisecond):
		s.Fail("Did not receive event in time")
	}

	// Unsubscribe
	registry.Unsubscribe(subscriber)
}

// TestStreamingListener tests the ABCI listener functionality
func (s *KeeperTestSuite) TestStreamingListener() {
	intake := s.App.AppKeepers.StreamKeeper.Intake()
	listener := types.NewStreamingListener(intake)
	s.Require().NotNil(listener)

	// Test ListenFinalizeBlock (should not error)
	err := listener.ListenFinalizeBlock(context.Background(), abci.RequestFinalizeBlock{}, abci.ResponseFinalizeBlock{})
	s.Require().NoError(err)

	// Test ListenCommit (should not error)
	err = listener.ListenCommit(context.Background(), abci.ResponseCommit{}, nil)
	s.Require().NoError(err)
}

// TestIntakeChannel tests the intake channel functionality
func (s *KeeperTestSuite) TestIntakeChannel() {
	intake := s.App.AppKeepers.StreamKeeper.Intake()
	s.Require().NotNil(intake)

	// Test sending an event to the intake channel
	event := types.StreamEvent{
		Module:      types.ModuleNameBank,
		EventType:   types.EventTypeBalanceChange,
		Address:     s.TestAccs[0].String(),
		Denom:       "ujuno",
		BlockHeight: s.Ctx.BlockHeight(),
	}

	// Send event (should not block)
	select {
	case intake <- event:
		// Success
	case <-time.After(100 * time.Millisecond):
		s.Fail("Failed to send event to intake channel")
	}
}
