package tx

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"cosmossdk.io/math"
	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	suite "github.com/stretchr/testify/suite"
)

type TxTestSuite struct {
	e2esuite.E2ETestSuite
}

func TestTxTestSuite(t *testing.T) {
	s := e2esuite.NewE2ETestSuite(
		e2esuite.Spec,
		e2esuite.TxCfg,
	)

	suite.Run(t, s)
}

// TestSendTxDecrease tests that the feemarket will decrease until it hits the min gas price
// when gas utilization is below the target block utilization.
func (s *TxTestSuite) TestSendTxDecrease() {
	s.T().Parallel()
	// get nodes
	nodes := s.Chain.Nodes()
	s.Require().True(len(nodes) > 0)

	params := s.QueryParams()

	defaultGasPrice := s.QueryDefaultGasPrice()
	gas := int64(200000)
	minBaseFee := sdk.NewDecCoinFromDec(defaultGasPrice.Denom, defaultGasPrice.Amount.Mul(math.LegacyNewDec(gas)))
	minBaseFeeCoins := sdk.NewCoins(sdk.NewCoin(minBaseFee.Denom, minBaseFee.Amount.TruncateInt()))
	sendAmt := int64(100)

	s.Run("expect fee market state to decrease", func() {
		s.T().Log("performing sends...")
		sig := make(chan struct{})
		quit := make(chan struct{})
		defer close(quit)

		checkPrice := func(c, quit chan struct{}) {
			select {
			case <-time.After(500 * time.Millisecond):
				gasPrice := s.QueryDefaultGasPrice()
				s.T().Log("gas price", gasPrice.String())

				if gasPrice.Amount.Equal(params.MinBaseGasPrice) {
					c <- struct{}{}
				}
			case <-quit:
				return
			}
		}
		go checkPrice(sig, quit)

		select {
		case <-sig:
			break

		case <-time.After(100 * time.Millisecond):
			wg := &sync.WaitGroup{}
			wg.Add(3)

			smallSend := func(wg *sync.WaitGroup, userA, userB ibc.Wallet) {
				defer wg.Done()
				txResp, err := s.SendCoinsMultiBroadcast(
					context.Background(),
					userA,
					userB,
					sdk.NewCoins(sdk.NewCoin(s.Chain.Config().Denom, math.NewInt(sendAmt))),
					minBaseFeeCoins,
					gas,
					s.TxConfig.SmallSendsNum,
				)
				if err != nil {
					s.T().Log(err)
				} else if txResp != nil && txResp.CheckTx.Code != 0 {
					s.T().Log(txResp.CheckTx)
				}
			}

			go smallSend(wg, s.User1, s.User2)
			go smallSend(wg, s.User3, s.User2)
			go smallSend(wg, s.User2, s.User1)

			wg.Wait()
		}

		// wait for 5 blocks
		// query height
		height, err := s.Chain.Height(context.Background())
		s.Require().NoError(err)
		s.WaitForHeight(s.Chain, height+5)

		gasPrice := s.QueryDefaultGasPrice()
		s.T().Log("gas price", gasPrice.String())

		amt, err := s.Chain.GetBalance(context.Background(), s.User1.FormattedAddress(), minBaseFee.Denom)
		s.Require().NoError(err)
		s.Require().True(amt.LT(math.NewInt(e2esuite.InitBalance)), amt)
		s.T().Log("balance:", amt.String())
	})
}

// TestSendTxIncrease tests that the feemarket will increase
// when gas utilization is above the target block utilization.
func (s *TxTestSuite) TestSendTxIncrease() {
	// get nodes
	nodes := s.Chain.Nodes()
	s.Require().True(len(nodes) > 0)

	params := s.QueryParams()

	gas := int64(params.MaxBlockUtilization)
	sendAmt := int64(100)

	s.Run("expect fee market gas price to increase", func() {
		s.T().Log("performing sends...")
		sig := make(chan struct{})
		quit := make(chan struct{})
		defer close(quit)

		checkPrice := func(c, quit chan struct{}) {
			select {
			case <-time.After(500 * time.Millisecond):
				gasPrice := s.QueryDefaultGasPrice()
				s.T().Log("gas price", gasPrice.String())

				if gasPrice.Amount.GT(s.TxConfig.TargetIncreaseGasPrice) {
					c <- struct{}{}
				}
			case <-quit:
				return
			}
		}
		go checkPrice(sig, quit)

		select {
		case <-sig:
			break

		case <-time.After(100 * time.Millisecond):
			// send with the exact expected baseGasPrice
			baseGasPrice := s.QueryDefaultGasPrice()
			minBaseFee := sdk.NewDecCoinFromDec(baseGasPrice.Denom, baseGasPrice.Amount.Mul(math.LegacyNewDec(gas)))
			// add headroom
			minBaseFeeCoins := sdk.NewCoins(sdk.NewCoin(minBaseFee.Denom, minBaseFee.Amount.Add(math.LegacyNewDec(10)).TruncateInt()))

			wg := &sync.WaitGroup{}
			wg.Add(3)

			largeSend := func(wg *sync.WaitGroup, userA, userB ibc.Wallet) {
				defer wg.Done()
				txResp, err := s.SendCoinsMultiBroadcast(
					context.Background(),
					userA,
					userB,
					sdk.NewCoins(sdk.NewCoin(s.Chain.Config().Denom, math.NewInt(sendAmt))),
					minBaseFeeCoins,
					gas,
					s.TxConfig.LargeSendsNum,
				)
				if err != nil {
					s.T().Log(err)
				} else if txResp != nil && txResp.CheckTx.Code != 0 {
					s.T().Log(txResp.CheckTx)
				}
			}

			go largeSend(wg, s.User1, s.User2)
			go largeSend(wg, s.User3, s.User2)
			go largeSend(wg, s.User2, s.User1)

			wg.Wait()
		}

		// wait for 5 blocks
		// query height
		height, err := s.Chain.Height(context.Background())
		s.Require().NoError(err)
		s.WaitForHeight(s.Chain, height+5)

		gasPrice := s.QueryDefaultGasPrice()
		s.T().Log("gas price", gasPrice.String())

		amt, err := s.Chain.GetBalance(context.Background(), s.User1.FormattedAddress(), gasPrice.Denom)
		s.Require().NoError(err)
		s.T().Log("balance:", amt.String())
	})
}

func (s *TxTestSuite) TestSendTxFailures() {
	sendAmt := int64(100)
	gas := int64(200000)

	s.Run("submit tx with no gas attached", func() {
		// send one tx with no  gas or fee attached
		txResp, err := s.SendCoinsMultiBroadcast(
			context.Background(),
			s.User1,
			s.User3,
			sdk.NewCoins(sdk.NewCoin(s.Chain.Config().Denom, math.NewInt(sendAmt))),
			sdk.NewCoins(),
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
			context.Background(),
			s.User1,
			s.User3,
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
		balance := s.QueryBalance(s.User3)

		// send one tx with no  gas or fee attached
		txResp, err := s.SendCoinsMultiBroadcast(
			context.Background(),
			s.User3,
			s.User1,
			sdk.NewCoins(balance),
			sdk.NewCoins(balance),
			gas,
			1,
		)
		s.Require().NoError(err)
		s.Require().NotNil(txResp)
		s.Require().True(txResp.CheckTx.Code == 0)
		s.Require().True(txResp.TxResult.Code != 0)
		s.T().Log(txResp.TxResult.Log)
		s.Require().Contains(txResp.TxResult.Log, "insufficient funds")

		// ensure that balance is deducted for any tx passing checkTx
		newBalance := s.QueryBalance(s.User3)
		s.Require().True(newBalance.IsLT(balance), fmt.Sprintf("new balance: %d, original balance: %d",
			balance.Amount.Int64(),
			newBalance.Amount.Int64()))
	})

	s.Run("submit a tx for full balance - fail tx", func() {
		balance := s.QueryBalance(s.User1)

		defaultGasPrice := s.QueryDefaultGasPrice()
		minBaseFee := sdk.NewDecCoinFromDec(defaultGasPrice.Denom, defaultGasPrice.Amount.Mul(math.LegacyNewDec(gas)))
		minBaseFeeCoins := sdk.NewCoins(sdk.NewCoin(minBaseFee.Denom, minBaseFee.Amount.TruncateInt().Add(math.
			NewInt(100))))
		txResp, err := s.SendCoinsMultiBroadcast(
			context.Background(),
			s.User1,
			s.User3,
			sdk.NewCoins(balance),
			minBaseFeeCoins,
			gas,
			1,
		)
		s.Require().NoError(err)
		s.Require().NotNil(txResp)
		s.Require().True(txResp.CheckTx.Code == 0)
		s.Require().True(txResp.TxResult.Code != 0)
		s.T().Log(txResp.TxResult.Log)
		s.Require().Contains(txResp.TxResult.Log, "insufficient funds")

		// ensure that balance is deducted for any tx passing checkTx
		newBalance := s.QueryBalance(s.User3)
		s.Require().True(newBalance.IsLT(balance), fmt.Sprintf("new balance: %d, original balance: %d",
			balance.Amount.Int64(),
			newBalance.Amount.Int64()))
	})

	s.Run("submit a tx with fee greater than full balance - fail checktx", func() {
		balance := s.QueryBalance(s.User1)
		txResp, err := s.SendCoinsMultiBroadcast(
			context.Background(),
			s.User1,
			s.User3,
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
		newBalance := s.QueryBalance(s.User1)
		s.Require().True(newBalance.Equal(balance), fmt.Sprintf("new balance: %d, original balance: %d",
			balance.Amount.Int64(),
			newBalance.Amount.Int64()))
	})

	s.Run("submit 2 tx in the same block - fail checktx in 2nd", func() {
		balance := s.QueryBalance(s.User2)

		defaultGasPrice := s.QueryDefaultGasPrice()
		minBaseFee := sdk.NewDecCoinFromDec(defaultGasPrice.Denom, defaultGasPrice.Amount.Mul(math.LegacyNewDec(gas)))
		minBaseFeeCoins := sdk.NewCoins(sdk.NewCoin(minBaseFee.Denom, minBaseFee.Amount.TruncateInt().Add(math.
			NewInt(100))))
		// send one tx with no  gas or fee attached
		txResp, err := s.SendCoinsMultiBroadcastAsync(
			context.Background(),
			s.User2,
			s.User1,
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
			context.Background(),
			s.User2,
			s.User1,
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

		// wait for 5 blocks
		// query height
		height, err := s.Chain.Height(context.Background())
		s.Require().NoError(err)
		s.WaitForHeight(s.Chain, height+5)

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
		s.User2 = s.GetAndFundTestUser(context.Background(), s.T().Name(), 200000000000, s.Chain)
	})
}
