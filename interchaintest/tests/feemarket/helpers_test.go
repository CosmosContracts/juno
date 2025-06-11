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
	messagesPerTransaction := 100       // 50 * 70k = 3.5M gas per transaction - exceeds target
	gasPerTransaction := int64(4000000) // High gas limit to accommodate 50 messages

	s.T().Logf("Congesting network: %d users, %d msgs per tx (~3.5M actual gas per tx)",
		len(users), messagesPerTransaction)

	// Get current gas price and calculate fees
	currentGasPrice := s.QueryFeemarketGasPrice(s.Denom)
	feeMultiplier := math.LegacyNewDec(3) // High multiplier to ensure inclusion
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
				messagesPerTransaction,
			)

			txResp, err = s.SendCoinsMultiBroadcast(
				from,
				receiver,
				sdk.NewCoins(sdk.NewCoin(s.Chain.Config().Denom, math.NewInt(sendAmt))),
				fees,
				gasPerTransaction,
				messagesPerTransaction,
			)

			txResp, err = s.SendCoinsMultiBroadcast(
				from,
				receiver,
				sdk.NewCoins(sdk.NewCoin(s.Chain.Config().Denom, math.NewInt(sendAmt))),
				fees,
				gasPerTransaction,
				messagesPerTransaction,
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
