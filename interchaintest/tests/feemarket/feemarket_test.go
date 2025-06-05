package feemarket_test

import (
	"context"
	"fmt"
	"testing"
	"time"

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

func (s *FeemarketTestSuite) TestQueryParams() {
	require := s.Require()
	s.Run("query params", func() {
		// query params
		params := s.QueryFeemarketParams()

		s.T().Logf("feemarket params: %s", params.String())

		// expect validate to pass
		require.NoError(params.ValidateBasic(), params)
	})
}

func (s *FeemarketTestSuite) TestQueryState() {
	require := s.Require()
	s.Run("query state", func() {
		// query state
		state := s.QueryFeemarketState()

		s.T().Logf("feemarket state: %s", state.String())

		// expect validate to pass
		require.NoError(state.ValidateBasic(), state)
	})
}

func (s *FeemarketTestSuite) TestQueryGasPrice() {
	require := s.Require()
	s.Run("query gas price", func() {
		// query gas price
		gasPrice := s.QueryFeemarketGasPrice(s.Denom)

		s.T().Logf("feemarket gas price: %s", gasPrice.String())

		// expect validate to pass
		require.NoError(gasPrice.Validate(), gasPrice)
	})
}

// TestSendTxDecrease tests that the feemarket will decrease until it hits the min gas price
// when gas utilization is below the target block utilization.
func (s *FeemarketTestSuite) TestSendTxDecrease() {
	nodes := s.Chain.Nodes()
	s.Require().True(len(nodes) > 0)

	params := s.QueryFeemarketParams()

	// First, we need to ensure gas prices are elevated
	s.T().Log("Setting up elevated gas prices...")
	s.elevateGasPrices(params)

	// Record initial elevated state
	initialGasPrice := s.QueryFeemarketGasPrice(s.Denom)
	s.T().Logf("Initial elevated gas price: %s", initialGasPrice.String())
	s.T().Logf("Target minimum gas price: %s", params.MinBaseGasPrice.String())

	// Ensure we're starting from an elevated price
	s.Require().True(initialGasPrice.Amount.GT(params.MinBaseGasPrice),
		"Gas price not elevated. Initial: %s, Min: %s",
		initialGasPrice.String(), params.MinBaseGasPrice.String())

	// Setup test users
	users := []ibc.Wallet{
		s.GetAndFundTestUser("user1", 200000000000, s.Chain),
		s.GetAndFundTestUser("user2", 200000000000, s.Chain),
		s.GetAndFundTestUser("user3", 200000000000, s.Chain),
	}

	// Use reasonable gas that won't congest the network
	gasPerTx := int64(200000)
	sendAmt := int64(100)

	// Wait for chain to stabilize
	err := testutil.WaitForBlocks(s.Ctx, 1, s.Chain)
	s.Require().NoError(err)

	// Monitor gas prices
	priceUpdates := make(chan sdk.DecCoin, 100)
	stopMonitoring := make(chan struct{})
	monitoringDone := make(chan struct{})

	go s.monitorGasPrice(priceUpdates, stopMonitoring, monitoringDone)

	// Send minimal transactions to allow price to decrease
	txErrors := s.sendMinimalTransactions(users, gasPerTx, sendAmt)

	// Wait for multiple blocks to allow price decay
	currentHeight, err := s.Chain.Height(s.Ctx)
	s.Require().NoError(err)
	s.WaitForHeight(s.Chain, currentHeight+5)

	// Additional wait to capture more price movements
	time.Sleep(2 * time.Second)

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
	s.Require().True(len(txErrors) == 0, "Transactions failed: %v", txErrors)

	// Check that gas price decreased
	finalGasPrice := s.QueryFeemarketGasPrice(s.Denom)
	s.T().Logf("Final gas price: %s", finalGasPrice.String())

	s.Require().True(finalGasPrice.Amount.LTE(initialGasPrice.Amount),
		"Gas price did not decrease. Initial: %s, Final: %s",
		initialGasPrice.String(), finalGasPrice.String())

	// Verify it reached the minimum
	s.Require().True(finalGasPrice.Amount.Equal(params.MinBaseGasPrice),
		"Gas price did not reach minimum. Current: %s, Min: %s",
		finalGasPrice.String(), params.MinBaseGasPrice.String())

	// Verify price trend was decreasing
	s.verifyPriceDecreaseTrend(prices, params.MinBaseGasPrice)

	// Verify user balances decreased (they paid for transactions)
	for _, user := range users {
		amt, err := s.Chain.GetBalance(s.Ctx, user.FormattedAddress(), s.Denom)
		s.Require().NoError(err)
		s.Require().True(amt.LT(math.NewInt(e2esuite.InitBalance)),
			"User balance did not decrease: %s", amt.String())
	}
}

// TestSendTxIncrease tests that the feemarket will increase
// when gas utilization is above the target block utilization.
func (s *FeemarketTestSuite) TestSendTxIncrease() {
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
	}

	// Monitor gas prices in separate goroutine
	priceUpdates := make(chan sdk.DecCoin, 100)
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

	// Check that gas price increased during congestion (but may have decreased afterward)
	finalGasPrice := s.QueryFeemarketGasPrice(s.Denom)
	s.T().Logf("Final gas price: %s", finalGasPrice.String())
	s.T().Logf("Maximum gas price during congestion: %s%s", maxGasPrice.String(), s.Denom)

	s.Require().True(maxGasPrice.GT(initialGasPrice.Amount),
		"Gas price did not increase during congestion. Initial: %s, Maximum: %s%s",
		initialGasPrice.String(), maxGasPrice.String(), s.Denom)

	// Verify there was indeed an increase (as a percentage)
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
	user3 := s.GetAndFundTestUser("user3", 200000000000, s.Chain)

	err := testutil.WaitForBlocks(s.Ctx, 5, s.Chain)
	require.NoError(s.T(), err)

	s.Run("submit tx with no gas attached", func() {
		// send one tx with no  gas or fee attached
		txResp, err := s.SendCoinsMultiBroadcast(
			user1,
			user3,
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
			user3,
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
		balance := s.QueryBankBalance(user3)

		txResp, err := s.SendCoinsMultiBroadcast(
			user3,
			user1,
			sdk.NewCoins(balance),
			sdk.NewCoins(balance),
			gas,
			1,
		)
		s.Require().NoError(err)
		s.Require().NotNil(txResp)
		s.Require().True(txResp.CheckTx.Code != 0)
		s.T().Log(txResp.CheckTx.Log)
		s.Require().Contains(txResp.CheckTx.Log, "insufficient funds")

		// ensure that balance is deducted for any tx passing checkTx
		newBalance := s.QueryBankBalance(user3)
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
			user3,
			sdk.NewCoins(balance),
			minBaseFeeCoins,
			gas,
			1,
		)
		s.Require().NoError(err)
		s.Require().NotNil(txResp)
		s.Require().True(txResp.CheckTx.Code != 0)
		s.T().Log(txResp.CheckTx.Log)
		s.Require().Contains(txResp.CheckTx.Log, "insufficient funds")

		// ensure that balance is deducted for any tx passing checkTx
		newBalance := s.QueryBankBalance(user3)
		s.Require().True(newBalance.IsLT(balance), fmt.Sprintf("new balance: %d, original balance: %d",
			balance.Amount.Int64(),
			newBalance.Amount.Int64()))
	})

	s.Run("submit a tx with fee greater than full balance - fail checktx", func() {
		balance := s.QueryBankBalance(user1)
		txResp, err := s.SendCoinsMultiBroadcast(
			user1,
			user3,
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

	s.Run("submit 2 tx in the same block - fail checktx in 2nd", func() {
		balance := s.QueryBankBalance(user2)

		defaultGasPrice := s.QueryFeemarketGasPrice(s.Denom)
		minBaseFee := sdk.NewDecCoinFromDec(defaultGasPrice.Denom, defaultGasPrice.Amount.Mul(math.LegacyNewDec(gas)))
		minBaseFeeCoins := sdk.NewCoins(sdk.NewCoin(minBaseFee.Denom, minBaseFee.Amount.TruncateInt().Add(math.
			NewInt(100))))
		txResp, err := s.SendCoinsMultiBroadcastAsync(
			user2,
			user1,
			sdk.NewCoins(balance.SubAmount(minBaseFeeCoins.AmountOf(minBaseFee.Denom))),
			minBaseFeeCoins,
			gas,
			1,
			false,
		)
		s.Require().NoError(err)
		s.Require().NotNil(txResp)
		s.Require().True(txResp.Code == 0)
		hash1 := txResp.Hash

		txResp, err = s.SendCoinsMultiBroadcastAsync(
			user2,
			user1,
			minBaseFeeCoins,
			minBaseFeeCoins,
			gas,
			1,
			true,
		)
		s.Require().NoError(err)
		s.Require().NotNil(txResp)
		s.Require().True(txResp.Code == 0)
		hash2 := txResp.Hash

		nodes := s.Chain.Nodes()
		s.Require().True(len(nodes) > 0)

		// wait for 1 block
		// query height
		height, err := s.Chain.Height(context.Background())
		s.Require().NoError(err)
		s.WaitForHeight(s.Chain, height+1)

		// after waiting, we can now query the Tx Responses
		resp, err := nodes[0].TxHashToResponse(context.Background(), hash1.String())
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Require().True(resp.Code == 0)

		resp, err = nodes[0].TxHashToResponse(context.Background(), hash2.String())
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Require().True(resp.Code != 0)
		s.Require().Contains(resp.RawLog, "error escrowing funds")
		s.T().Log(resp.RawLog)

		// reset the users and balances
		user1 = s.GetAndFundTestUser("user1", 200000000000, s.Chain)
		user2 = s.GetAndFundTestUser("user2", 200000000000, s.Chain)
		user3 = s.GetAndFundTestUser("user3", 200000000000, s.Chain)
	})
}
