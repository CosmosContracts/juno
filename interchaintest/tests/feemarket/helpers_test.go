package feemarket_test

import (
	"fmt"
	"sync"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"

	feemarketypes "github.com/CosmosContracts/juno/v30/x/feemarket/types"
)

// waitForMinimumGasPrice waits for the gas price to reach the minimum by sending minimal transactions
func (s *FeemarketTestSuite) waitForMinimumGasPrice(params feemarketypes.Params) {
	// Setup temporary users for this operation
	// tempUsers := []ibc.Wallet{
	// 	s.GetAndFundTestUser("temp1", 200000000000, s.Chain),
	// 	s.GetAndFundTestUser("temp2", 200000000000, s.Chain),
	// 	s.GetAndFundTestUser("temp3", 200000000000, s.Chain),
	// }

	// gasPerTx := int64(200000) // Use minimal gas
	// sendAmt := int64(100)
	maxAttempts := 10
	attempt := 0

	for attempt < maxAttempts {
		currentGasPrice := s.QueryFeemarketGasPrice(s.Denom)
		s.T().Logf("Attempt %d: Current gas price: %s, Target min: %s",
			attempt+1, currentGasPrice.String(), params.MinBaseGasPrice.String())

		// If we've reached the minimum, we're done
		if currentGasPrice.Amount.Equal(params.MinBaseGasPrice) {
			s.T().Log("Gas price has reached minimum")
			return
		}

		// // Send minimal transactions to allow price to decrease
		// txErrors := s.sendMinimalTransactions(tempUsers, gasPerTx, sendAmt)
		// if len(txErrors) > 0 {
		// 	s.T().Logf("Some transactions failed during gas price normalization: %v", txErrors)
		// }

		// Wait for a few blocks to allow the price to adjust
		currentHeight, err := s.Chain.Height(s.Ctx)
		s.Require().NoError(err)
		s.WaitForHeight(s.Chain, currentHeight+2)

		attempt++
	}

	finalGasPrice := s.QueryFeemarketGasPrice(s.Denom)
	if !finalGasPrice.Amount.Equal(params.MinBaseGasPrice) {
		s.T().Logf("Warning: Gas price did not fully reach minimum after %d attempts. Current: %s, Min: %s",
			maxAttempts, finalGasPrice.String(), params.MinBaseGasPrice.String())

		// Wait for chain to stabilize anyway
		err := testutil.WaitForBlocks(s.Ctx, 2, s.Chain)
		s.Require().NoError(err)
	}
}

// elevateGasPrices creates congestion to raise gas prices before testing decrease
func (s *FeemarketTestSuite) elevateGasPrices(params feemarketypes.Params) {
	tempUsers := []ibc.Wallet{
		s.GetAndFundTestUser("temp1", 200000000000, s.Chain),
		s.GetAndFundTestUser("temp2", 200000000000, s.Chain),
	}

	gasPerTx := int64(params.MaxBlockUtilization / 3)
	iterations := 3

	for range iterations {
		var wg sync.WaitGroup
		baseGasPrice := s.QueryFeemarketGasPrice(s.Denom)

		// Calculate fees with buffer
		feeAmount := baseGasPrice.Amount.Mul(math.LegacyNewDec(gasPerTx)).Mul(math.LegacyNewDec(2))
		fees := sdk.NewCoins(sdk.NewCoin(baseGasPrice.Denom, feeAmount.TruncateInt()))

		// Send large transactions
		for _, user := range tempUsers {
			wg.Add(1)
			go func(sender ibc.Wallet) {
				defer wg.Done()

				_, _ = s.SendCoinsMultiBroadcast(
					sender,
					tempUsers[0], // just send to first user
					sdk.NewCoins(sdk.NewCoin(s.Chain.Config().Denom, math.NewInt(100))),
					fees,
					gasPerTx,
					10, // send 10 transactions
				)
			}(user)
		}

		wg.Wait()

		// Wait for block
		height, _ := s.Chain.Height(s.Ctx)
		s.WaitForHeight(s.Chain, height+1)
	}

	// Ensure price is elevated
	currentPrice := s.QueryFeemarketGasPrice(s.Denom)
	if currentPrice.Amount.LTE(params.MinBaseGasPrice) {
		s.T().Fatal("Failed to elevate gas prices for decrease test")
	}
}

// sendMinimalTransactions sends a small number of transactions to avoid congestion
func (s *FeemarketTestSuite) sendMinimalTransactions(
	users []ibc.Wallet,
	gasPerTx int64,
	sendAmt int64,
) []error {
	var (
		wg       sync.WaitGroup
		errorsMu sync.Mutex
		errors   []error
	)

	baseGasPrice := s.QueryFeemarketGasPrice(s.Denom)

	// Calculate fees - use current price without much buffer
	feeAmount := baseGasPrice.Amount.
		Mul(math.LegacyNewDec(gasPerTx))
		// Mul(math.LegacyMustNewDecFromStr("1.1"))
	fees := sdk.NewCoins(sdk.NewCoin(baseGasPrice.Denom, feeAmount.TruncateInt()))

	// Send minimal transactions - one per user
	for i, sender := range users {
		receiver := users[(i+1)%len(users)]
		wg.Add(1)

		go func(from, to ibc.Wallet) {
			defer wg.Done()

			// Send only SmallSendsNum transactions
			txResp, err := s.SendCoinsMultiBroadcast(
				from,
				to,
				sdk.NewCoins(sdk.NewCoin(s.Chain.Config().Denom, math.NewInt(sendAmt))),
				fees,
				gasPerTx,
				s.TxConfig.SmallSendsNum,
			)

			if err != nil {
				errorsMu.Lock()
				errors = append(errors, fmt.Errorf("broadcast error: %w", err))
				errorsMu.Unlock()
				return
			}

			if txResp != nil && txResp.CheckTx.Code != 0 {
				errorsMu.Lock()
				errors = append(errors, fmt.Errorf("check tx failed: %s", txResp.CheckTx.Log))
				errorsMu.Unlock()
			}
		}(sender, receiver)
	}

	wg.Wait()
	return errors
}

// verifyPriceDecreaseTrend ensures prices generally decreased over time
func (s *FeemarketTestSuite) verifyPriceDecreaseTrend(prices []sdk.DecCoin, minPrice math.LegacyDec) {
	if len(prices) < 2 {
		return
	}

	// Track how many samples showed decrease
	decreases := 0
	hitMinimum := false

	for i := 1; i < len(prices); i++ {
		if prices[i].Amount.LT(prices[i-1].Amount) {
			decreases++
		}
		if prices[i].Amount.Equal(minPrice) {
			hitMinimum = true
		}
	}

	trendPercentage := float64(decreases) / float64(len(prices)-1) * 100
	s.T().Logf("Price decrease trend: %.2f%% of samples showed decrease", trendPercentage)
	s.T().Logf("Hit minimum price: %v", hitMinimum)

	// At least 50% of samples should show a decreasing trend
	// (lower threshold than increase test because prices may stabilize at minimum)
	s.Require().True(trendPercentage >= 50.0 || hitMinimum,
		"Price trend not sufficiently decreasing: %.2f%% and didn't hit minimum", trendPercentage)
}

// monitorGasPrice continuously monitors gas price changes
func (s *FeemarketTestSuite) monitorGasPrice(
	updates chan<- sdk.DecCoin,
	stop <-chan struct{},
	done chan<- struct{},
) {
	defer close(done)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			gasPrice := s.QueryFeemarketGasPrice(s.Denom)
			select {
			case updates <- gasPrice:
			case <-stop:
				return
			}
		}
	}
}

func (s *FeemarketTestSuite) createNetworkCongestion(
	users []ibc.Wallet,
	sendAmt int64,
) []error {
	var allErrors []error

	// Simple approach: send very few transactions but with many messages each
	// Goal: ensure we exceed MaxBlockUtilization (3M gas) per block with actual gas consumption
	s.T().Logf("Using %d users for network congestion", len(users))

	// Each bank send consumes ~70k gas
	// To exceed 3M gas target, we need: 3,000,000 / 70,000 = ~43 bank sends minimum
	// Let's use 50 messages per transaction to be safe, and send one transaction per user
	messagesPerTransaction := 50        // 50 * 70k = 3.5M gas per transaction - exceeds target
	gasPerTransaction := int64(4000000) // High gas limit to accommodate 50 messages

	s.T().Logf("Congesting network: %d users, %d msgs per tx (~3.5M actual gas per tx)",
		len(users), messagesPerTransaction)

	// Get current gas price and calculate fees
	currentGasPrice := s.QueryFeemarketGasPrice(s.Denom)
	feeMultiplier := math.LegacyNewDec(5) // High multiplier to ensure inclusion
	feeAmount := currentGasPrice.Amount.Mul(math.LegacyNewDec(gasPerTransaction)).Mul(feeMultiplier)
	fees := sdk.NewCoins(sdk.NewCoin(currentGasPrice.Denom, feeAmount.TruncateInt()))

	var wg sync.WaitGroup
	var roundErrors []error
	var errorsMu sync.Mutex
	successCount := 0

	// Send one big transaction per user simultaneously
	for userIdx, sender := range users {
		wg.Add(1)

		go func(from ibc.Wallet, userIndex int) {
			defer wg.Done()

			// Send to a different user in round-robin fashion
			receiver := users[(userIndex+1)%len(users)]

			s.T().Logf("User %d sending transaction with %d messages", userIndex, messagesPerTransaction)

			txResp, err := s.SendCoinsMultiBroadcast(
				from,
				receiver,
				sdk.NewCoins(sdk.NewCoin(s.Chain.Config().Denom, math.NewInt(sendAmt))),
				fees,
				gasPerTransaction,
				messagesPerTransaction, // Many messages in single transaction
			)

			if err != nil {
				errorsMu.Lock()
				roundErrors = append(roundErrors,
					fmt.Errorf("user %d: broadcast error: %w", userIndex, err))
				errorsMu.Unlock()
				return
			}

			if txResp != nil && txResp.CheckTx.Code != 0 {
				errorsMu.Lock()
				roundErrors = append(roundErrors,
					fmt.Errorf("user %d: broadcast failed with code %d: %s",
						userIndex, txResp.CheckTx.Code, txResp.CheckTx.Log))
				errorsMu.Unlock()
			} else {
				errorsMu.Lock()
				successCount++
				errorsMu.Unlock()
				s.T().Logf("User %d transaction successful", userIndex)
			}
		}(sender, userIdx)
	}

	wg.Wait()

	// Wait for transactions to be included in blocks
	currentHeight, _ := s.Chain.Height(s.Ctx)
	s.WaitForHeight(s.Chain, currentHeight+2)

	totalMessages := successCount * messagesPerTransaction
	expectedGasConsumption := totalMessages * 70000 // ~70k gas per bank send
	s.T().Logf("Sent %d/%d successful transactions (%d total messages, ~%d gas consumption)",
		successCount, len(users), totalMessages, expectedGasConsumption)

	if len(roundErrors) > 0 {
		s.T().Logf("Transaction errors: %d", len(roundErrors))
		allErrors = append(allErrors, roundErrors...)
	}

	// Check gas price change
	newGasPrice := s.QueryFeemarketGasPrice(s.Denom)
	priceRatio := newGasPrice.Amount.Quo(currentGasPrice.Amount)
	s.T().Logf("Gas price after congestion: %s -> %s (%.2fx)",
		currentGasPrice.String(), newGasPrice.String(), priceRatio.MustFloat64())

	return allErrors
}

// verifyPriceTrend ensures prices generally increased over time
func (s *FeemarketTestSuite) verifyPriceTrend(prices []sdk.DecCoin) {
	if len(prices) < 4 { // Need at least 4 samples for windowing
		s.T().Logf("Not enough price samples (%d) for trend analysis", len(prices))
		return
	}

	// Calculate moving average to smooth out fluctuations
	windowSize := 3
	increases := 0

	// Start from windowSize+1 to ensure we have enough data for both windows
	for i := windowSize + 1; i < len(prices); i++ {
		currentAvg := s.calculateAverage(prices[i-windowSize : i])
		previousAvg := s.calculateAverage(prices[i-windowSize-1 : i-1])

		if currentAvg.GT(previousAvg) {
			increases++
		}
	}

	// Only calculate trend if we have samples to analyze
	totalComparisons := len(prices) - windowSize - 1
	if totalComparisons <= 0 {
		s.T().Logf("Not enough price samples for trend comparison")
		return
	}

	trendPercentage := float64(increases) / float64(totalComparisons) * 100
	s.T().Logf("Price increase trend: %.2f%% of samples showed increase (%d/%d)",
		trendPercentage, increases, totalComparisons)

	// At least 30% of samples should show an increasing trend
	// (lowered threshold since we expect prices to decrease after congestion ends)
	s.Require().True(trendPercentage >= 30.0,
		"Price trend not sufficiently increasing: %.2f%%", trendPercentage)
}

// calculateAverage computes the average of a slice of DecCoins
func (s *FeemarketTestSuite) calculateAverage(prices []sdk.DecCoin) math.LegacyDec {
	sum := math.LegacyZeroDec()
	for _, p := range prices {
		sum = sum.Add(p.Amount)
	}
	return sum.Quo(math.LegacyNewDec(int64(len(prices))))
}
