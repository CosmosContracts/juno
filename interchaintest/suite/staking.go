package suite

import (
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *E2ETestSuite) StakeTokens(chain *cosmos.CosmosChain, user ibc.Wallet, valoper, coinAmt string, fees sdk.Coins, skipTxCheck bool) {
	t := s.T()
	// amount is #utoken
	cmd := []string{
		"staking", "delegate", valoper, coinAmt,
		"--from", user.KeyName(),
		"--gas", "auto",
		"--fees", fees.String(),
	}
	txHash, _ := s.ExecTx(s.Chain, user.KeyName(), false, true, cmd...)
	if skipTxCheck {
		if err := testutil.WaitForBlocks(s.Ctx, 1, chain); err != nil {
			t.Fatal(err)
		}
		return
	}
	// convert stdout into a TxResponse
	txRes, err := chain.GetTransaction(txHash)
	if err != nil {
		t.Fatal(err)
	}

	s.DebugOutput(string(txRes.RawLog))

	if err := testutil.WaitForBlocks(s.Ctx, 2, chain); err != nil {
		t.Fatal(err)
	}
}

func (s *E2ETestSuite) ClaimStakingRewards(chain *cosmos.CosmosChain, user ibc.Wallet, valoper string, fees sdk.Coins) {
	t := s.T()
	cmd := []string{
		"distribution", "withdraw-rewards", valoper,
		"--from", user.KeyName(),
		"--gas", "auto",
		"--fees", fees.String(),
	}
	txHash, _ := s.ExecTx(s.Chain, user.KeyName(), false, true, cmd...)
	// convert stdout into a TxResponse
	txRes, err := chain.GetTransaction(txHash)
	if err != nil {
		t.Fatal(err)
	}

	s.DebugOutput(string(txRes.RawLog))

	if err := testutil.WaitForBlocks(s.Ctx, 2, chain); err != nil {
		t.Fatal(err)
	}
}
