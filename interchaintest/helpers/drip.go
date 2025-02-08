package helpers

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
)

func DripTokens(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, coinAmt string) {
	// amount is #utoken
	cmd := []string{
		"junod", "tx", "drip", "distribute-tokens", user.FormattedAddress(), coinAmt,
		"--home", chain.HomeDir(),
		"--node", chain.GetRPCAddress(),
		"--chain-id", chain.Config().ChainID,
		"--gas", "500000",
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	}
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	debugOutput(t, string(stdout))

	if err := testutil.WaitForBlocks(ctx, 2, chain); err != nil {
		t.Fatal(err)
	}
}
