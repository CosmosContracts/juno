package suite

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
)

func (s *E2ETestSuite) DripTokens(chain *cosmos.CosmosChain, user ibc.Wallet, coinAmt string, fees sdk.Coins) {
	t := s.T()
	cmd := []string{
		"drip", "distribute-tokens", user.FormattedAddress(), coinAmt,
		"--gas", "auto",
		"--fees", fees.String(),
	}
	txHash, _ := s.ExecTx(s.Chain, user.KeyName(), false, false, cmd...)
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
