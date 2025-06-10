package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/CosmosContracts/juno/v30/testutil"
	"github.com/CosmosContracts/juno/v30/x/stream/types"
)

type IntegrationTestSuite struct {
	testutil.KeeperTestHelper
}

func (s *IntegrationTestSuite) SetupTest() {
	s.Setup()
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// TestBankTransactionEvent tests that bank transactions work with stream module enabled
func (s *IntegrationTestSuite) TestBankTransactionEvent() {
	// Fund the test account
	sender := s.TestAccs[0]
	receiver := s.TestAccs[1]
	amount := sdk.NewCoins(sdk.NewCoin("ujuno", math.NewInt(1000000)))

	// Fund sender account
	err := s.App.AppKeepers.BankKeeper.MintCoins(s.Ctx, types.ModuleName, amount)
	s.Require().NoError(err)
	err = s.App.AppKeepers.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, types.ModuleName, sender, amount)
	s.Require().NoError(err)

	// Verify initial balances
	senderBalance := s.App.AppKeepers.BankKeeper.GetBalance(s.Ctx, sender, "ujuno")
	s.Require().Equal(math.NewInt(1000000), senderBalance.Amount)

	// Send the transaction
	sendAmount := sdk.NewCoins(sdk.NewCoin("ujuno", math.NewInt(100000)))
	err = s.App.AppKeepers.BankKeeper.SendCoins(s.Ctx, sender, receiver, sendAmount)
	s.Require().NoError(err)

	// Verify final balances
	senderBalance = s.App.AppKeepers.BankKeeper.GetBalance(s.Ctx, sender, "ujuno")
	receiverBalance := s.App.AppKeepers.BankKeeper.GetBalance(s.Ctx, receiver, "ujuno")

	s.Require().Equal(math.NewInt(900000), senderBalance.Amount)
	s.Require().Equal(math.NewInt(100000), receiverBalance.Amount)
}

// TestStreamModuleIntegration tests the stream module's integration with the app
func (s *IntegrationTestSuite) TestStreamModuleIntegration() {
	// Verify the stream keeper is properly initialized
	s.Require().NotNil(s.App.AppKeepers.StreamKeeper)

	// Verify the intake channel exists
	intake := s.App.AppKeepers.StreamKeeper.Intake()
	s.Require().NotNil(intake)

	// Test that we can send an event to the intake (won't block)
	testEvent := types.StreamEvent{
		Module:      types.ModuleNameBank,
		EventType:   types.EventTypeBalanceChange,
		Address:     s.TestAccs[0].String(),
		Denom:       "ujuno",
		BlockHeight: s.Ctx.BlockHeight(),
	}

	// This should not panic or block
	select {
	case intake <- testEvent:
		// Success
	default:
		s.Fail("Intake channel should not be full")
	}
}
