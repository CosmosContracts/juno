package feemarket_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"
)

type FeemarketTestSuite struct {
	*e2esuite.E2ETestSuite
}

func TestFeemarketTestSuite(t *testing.T) {
	numValidators := 1
	numFullNodes := 0

	spec := &interchaintest.ChainSpec{
		ChainName:     "juno-fees",
		Name:          "juno",
		NumValidators: &numValidators,
		NumFullNodes:  &numFullNodes,
		Version:       e2esuite.DefaultSpec.Version,
		NoHostMount:   &e2esuite.DefaultNoHostMount,
		ChainConfig:   e2esuite.DefaultConfig,
	}

	s := e2esuite.NewE2ETestSuite(
		[]*interchaintest.ChainSpec{spec},
		e2esuite.DefaultTxCfg,
	)

	t.Parallel()
	t.Cleanup(func() {
		_ = s.Ic.Close()
	})

	testSuite := &FeemarketTestSuite{E2ETestSuite: s}
	suite.Run(t, testSuite)
}

func (s *FeemarketTestSuite) SetupSubTest() {
	height, err := s.Chain.Height(s.Ctx)
	s.Require().NoError(err)
	s.WaitForHeight(s.Chain, height+1)

	state := s.QueryFeemarketState()
	s.T().Log("state at block height", height+1, ":", state.String())
	gasPrice := s.QueryFeemarketGasPrice(s.Denom)
	s.T().Log("gas price at block height", height+1, ":", gasPrice.String())
}

// TestFeemarketUpdate tests that the feemarket will increase
// when gas utilization is above the target block utilization
// and that the gas price will decrease after the congestion ends
func (s *FeemarketTestSuite) TestFeemarketUpdate() {
	nodes := s.Chain.Nodes()
	s.Require().True(len(nodes) > 0)

	params := s.QueryFeemarketParams()
	sendAmt := int64(100)

	// Wait for gas price to reach minimum before starting the test
	s.T().Log("Waiting for gas price to reach minimum...")
	s.waitForMinimumGasPrice(params)

	// Record initial state
	initialGasPrice := s.QueryFeemarketGasPrice(s.Denom)
	s.T().Logf("Initial gas price: %s", initialGasPrice.String())

	// Ensure we're starting from the minimum gas price
	s.Require().True(initialGasPrice.Amount.Equal(params.MinBaseGasPrice),
		"Gas price not at minimum. Current: %s, Min: %s",
		initialGasPrice.String(), params.MinBaseGasPrice.String())

	// Setup test users
	users := []ibc.Wallet{
		s.GetAndFundTestUser("user1", 200000000000, s.Chain),
		s.GetAndFundTestUser("user2", 200000000000, s.Chain),
		s.GetAndFundTestUser("user3", 200000000000, s.Chain),
		s.GetAndFundTestUser("user4", 200000000000, s.Chain),
		s.GetAndFundTestUser("user5", 200000000000, s.Chain),
	}

	// Monitor gas prices in separate goroutine
	priceUpdates := make(chan sdk.DecCoin, 300)
	stopMonitoring := make(chan struct{})
	monitoringDone := make(chan struct{})

	go s.monitorGasPrice(priceUpdates, stopMonitoring, monitoringDone)

	// Send transactions to create congestion
	txErrors := s.createNetworkCongestion(users, sendAmt)

	if len(txErrors) > 0 {
		s.T().Logf("Some transactions failed during network congestion: %v", txErrors)
	}

	// Stop monitoring and collect results
	close(stopMonitoring)
	<-monitoringDone
	close(priceUpdates)

	// Analyze price changes
	prices := []sdk.DecCoin{}
	for price := range priceUpdates {
		prices = append(prices, price)
	}

	// Verify results
	s.Require().True(len(prices) > 0, "No price updates captured")
	s.T().Logf("Price updates: %v", prices)

	// Find the maximum gas price during the congestion period
	maxGasPrice := initialGasPrice.Amount
	for _, price := range prices {
		if price.Amount.GT(maxGasPrice) {
			maxGasPrice = price.Amount
		}
	}

	// Check that gas price increased during congestion
	finalGasPrice := s.QueryFeemarketGasPrice(s.Denom)
	s.T().Logf("Final gas price: %s", finalGasPrice.String())
	s.T().Logf("Maximum gas price during congestion: %s%s", maxGasPrice.String(), s.Denom)

	s.Require().True(maxGasPrice.GT(initialGasPrice.Amount),
		"Gas price did not increase during congestion. Initial: %s, Maximum: %s%s",
		initialGasPrice.String(), maxGasPrice.String(), s.Denom)

	// Verify there was indeed an increase
	priceIncreaseRatio := maxGasPrice.Quo(initialGasPrice.Amount)
	s.T().Logf("Maximum price increase: %.2fx", priceIncreaseRatio.MustFloat64())
	s.Require().True(priceIncreaseRatio.GT(math.LegacyMustNewDecFromStr("1.1")),
		"Gas price should have increased by at least 10%% during congestion. Ratio: %.2fx",
		priceIncreaseRatio.MustFloat64())

	s.T().Logf("✅ Feemarket test PASSED: Gas price successfully increased from %s to %s during congestion",
		initialGasPrice.String(), maxGasPrice.String()+s.Denom)
}

func (s *FeemarketTestSuite) TestSendTxFailures() {
	sendAmt := int64(100)
	gas := int64(200000)

	user1 := s.GetAndFundTestUser("user1", 200000000000, s.Chain)
	user2 := s.GetAndFundTestUser("user2", 200000000000, s.Chain)

	err := testutil.WaitForBlocks(s.Ctx, 5, s.Chain)
	require.NoError(s.T(), err)

	s.Run("submit tx with no gas attached", func() {
		// send one tx with no  gas or fee attached
		txResp, err := s.SendCoinsMultiBroadcast(
			user1,
			user2,
			sdk.NewCoins(sdk.NewCoin(s.Chain.Config().Denom, math.NewInt(sendAmt))),
			sdk.NewCoins(sdk.NewCoin(s.Chain.Config().Denom, math.NewInt(1_000_000))),
			0,
			1,
		)
		s.Require().NoError(err)
		s.Require().NotNil(txResp)
		s.Require().True(txResp.CheckTx.Code != 0)
		s.T().Log(txResp.CheckTx.Log)
		s.Require().Contains(txResp.CheckTx.Log, "out of gas")
	})

	s.Run("submit tx with no fee", func() {
		txResp, err := s.SendCoinsMultiBroadcast(
			user1,
			user2,
			sdk.NewCoins(sdk.NewCoin(s.Chain.Config().Denom, math.NewInt(sendAmt))),
			sdk.NewCoins(),
			gas,
			1,
		)
		s.Require().NoError(err)
		s.Require().NotNil(txResp)
		s.Require().True(txResp.CheckTx.Code != 0)
		s.T().Log(txResp.CheckTx.Log)
		s.Require().Contains(txResp.CheckTx.Log, "no fee coin provided")
	})

	s.Run("fail a tx that uses full balance in fee - fail tx", func() {
		balance := s.QueryBankBalance(user2)

		txResp, err := s.SendCoinsMultiBroadcast(
			user2,
			user1,
			sdk.NewCoins(balance),
			sdk.NewCoins(balance),
			gas,
			1,
		)
		s.Require().NoError(err)
		s.Require().NotNil(txResp)
		s.Require().True(txResp.TxResult.Code != 0)
		s.T().Log(txResp.TxResult.Log)
		s.Require().Contains(txResp.TxResult.Log, "insufficient funds")

		// ensure that balance is deducted for any tx passing checkTx
		newBalance := s.QueryBankBalance(user2)
		s.Require().True(newBalance.IsLT(balance), fmt.Sprintf("new balance: %d, original balance: %d",
			balance.Amount.Int64(),
			newBalance.Amount.Int64()))
	})

	s.Run("submit a tx for full balance - fail tx", func() {
		balance := s.QueryBankBalance(user1)

		defaultGasPrice := s.QueryFeemarketGasPrice(s.Denom)
		minBaseFee := sdk.NewDecCoinFromDec(defaultGasPrice.Denom, defaultGasPrice.Amount.Mul(math.LegacyNewDec(gas)))
		minBaseFeeCoins := sdk.NewCoins(sdk.NewCoin(minBaseFee.Denom, minBaseFee.Amount.TruncateInt().Add(math.
			NewInt(100))))
		txResp, err := s.SendCoinsMultiBroadcast(
			user1,
			user2,
			sdk.NewCoins(balance),
			minBaseFeeCoins,
			gas,
			1,
		)
		s.Require().NoError(err)
		s.Require().NotNil(txResp)
		s.T().Log(txResp.TxResult)
		s.Require().True(txResp.TxResult.Code != 0)
		s.T().Log(txResp.TxResult.Log)
		s.Require().Contains(txResp.TxResult.Log, "insufficient funds")

		// ensure that balance is deducted for any tx passing checkTx
		newBalance := s.QueryBankBalance(user2)
		s.Require().True(newBalance.IsLT(balance), fmt.Sprintf("new balance: %d, original balance: %d",
			balance.Amount.Int64(),
			newBalance.Amount.Int64()))
	})

	s.Run("submit a tx with fee greater than full balance - fail checktx", func() {
		balance := s.QueryBankBalance(user1)
		txResp, err := s.SendCoinsMultiBroadcast(
			user1,
			user2,
			sdk.NewCoins(sdk.NewCoin(s.Chain.Config().Denom, math.NewInt(sendAmt))),
			sdk.NewCoins(balance.AddAmount(math.NewInt(110000))),
			gas,
			1,
		)
		s.Require().NoError(err)
		s.Require().NotNil(txResp)
		s.Require().True(txResp.CheckTx.Code != 0)
		s.T().Log(txResp.CheckTx.Log)
		s.Require().Contains(txResp.CheckTx.Log, "error escrowing funds")

		// ensure that no balance is deducted for a tx failing checkTx
		newBalance := s.QueryBankBalance(user1)
		s.Require().True(newBalance.Equal(balance), fmt.Sprintf("new balance: %d, original balance: %d",
			balance.Amount.Int64(),
			newBalance.Amount.Int64()))
	})
}
