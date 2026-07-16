package suite

import (
    "github.com/cosmos/interchaintest/v10/chain/cosmos"
    "github.com/cosmos/interchaintest/v10/ibc"
    "github.com/cosmos/interchaintest/v10/testutil"

    sdk "github.com/cosmos/cosmos-sdk/types"
    coretypes "github.com/cometbft/cometbft/rpc/core/types"
    stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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
	txHash, err := s.ExecTx(s.Chain, user.KeyName(), true, skipTxCheck, cmd...)
	if err != nil {
		t.Fatal(err)
	}

	if skipTxCheck {
		if err := testutil.WaitForBlocks(s.Ctx, 1, chain); err != nil {
			t.Fatal(err)
		}
		return
	}

	txRes, err := chain.GetTransaction(txHash)
	if err != nil {
		t.Fatal(err)
	}
	s.DebugOutput(string(txRes.RawLog))
	if err := testutil.WaitForBlocks(s.Ctx, 1, chain); err != nil {
		t.Fatal(err)
	}
}

// StakeTokensAsync builds a staking delegate tx and broadcasts it asynchronously with a custom memo and explicit sequence.
// This allows spamming multiple txs across many accounts without waiting for inclusion while keeping per-account sequences valid.
func (s *E2ETestSuite) StakeTokensAsync(
    user ibc.Wallet,
    valoper string,
    amount sdk.Coin,
    fees sdk.Coins,
    gas int64,
    memo string,
    sequence uint64,
) (*coretypes.ResultBroadcastTx, error) {
    // construct msg
    msg := &stakingtypes.MsgDelegate{
        DelegatorAddress: user.FormattedAddress(),
        ValidatorAddress: valoper,
        Amount:           amount,
    }

    // build tx bytes with memo and explicit sequence and broadcast async
    tx := s.CreateTxWithMemoAndSequence(s.Chain, user, fees.String(), gas, memo, sequence, msg)
    c := s.Chain.Nodes()[0].Client
    return c.BroadcastTxAsync(s.Ctx, tx)
}

func (s *E2ETestSuite) ClaimStakingRewards(chain *cosmos.CosmosChain, user ibc.Wallet, valoper string, fees sdk.Coins) {
	t := s.T()
	cmd := []string{
		"distribution", "withdraw-rewards", valoper,
		"--from", user.KeyName(),
		"--gas", "auto",
		"--fees", fees.String(),
	}
	txHash, _ := s.ExecTx(s.Chain, user.KeyName(), true, false, cmd...)
	// convert stdout into a TxResponse
	txRes, err := chain.GetTransaction(txHash)
	if err != nil {
		t.Fatal(err)
	}

	s.DebugOutput(string(txRes.RawLog))

	if err := testutil.WaitForBlocks(s.Ctx, 1, chain); err != nil {
		t.Fatal(err)
	}
}
