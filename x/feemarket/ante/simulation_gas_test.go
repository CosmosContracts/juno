package ante_test

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestSimulationIncludesUserFeeEscrowGas guards wallet gas estimation. A
// simulation must execute the same fee-escrow bank writes as delivery; otherwise
// wallets apply their gas adjustment to an estimate that is systematically low.
func (s *AnteTestSuite) TestSimulationIncludesUserFeeEscrowGas() {
	payer := s.fullAccs[0]
	payerAddr := payer.Account.GetAddress()
	fee := sdk.NewCoins(sdk.NewInt64Coin("stake", 1_000_000))
	s.FundAcc(payerAddr, sdk.NewCoins(sdk.NewInt64Coin("stake", 2_000_000)))

	account := s.App.AppKeepers.AccountKeeper.GetAccount(s.Ctx, payerAddr)
	s.Require().NoError(account.SetPubKey(payer.Priv.PubKey()))
	s.App.AppKeepers.AccountKeeper.SetAccount(s.Ctx, account)
	makeTx := func(gas uint64, fees sdk.Coins) sdk.Tx {
		s.Require().NoError(s.TxBuilder.SetMsgs(NewTestMsg(payerAddr)))
		s.TxBuilder.SetFeeAmount(fees)
		s.TxBuilder.SetGasLimit(gas)

		tx, err := s.CreateTestTx(
			[]cryptotypes.PrivKey{payer.Priv},
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			s.Ctx.ChainID(),
		)
		s.Require().NoError(err)
		return tx
	}

	simCtx, _ := s.Ctx.CacheContext()
	// Keplr's simulateAndSend path submits an empty fee amount, then computes
	// the real fee only after receiving the gas estimate.
	simCtx, err := s.AnteHandler(simCtx, makeTx(0, nil), true)
	s.Require().NoError(err)
	simulatedGas := simCtx.GasMeter().GasConsumed()

	deliverCtx, _ := s.Ctx.CacheContext()
	deliverCtx, err = s.AnteHandler(deliverCtx, makeTx(1_000_000, fee), false)
	s.Require().NoError(err)
	deliveredGas := deliverCtx.GasMeter().GasConsumed()

	s.T().Logf("simulated gas: %d; delivered gas: %d", simulatedGas, deliveredGas)
	// The positive simulation probe may differ slightly from the final fee due
	// to encoded coin amount size, but it must include the escrow store work.
	s.Require().InDelta(deliveredGas, simulatedGas, 1_000)
}
