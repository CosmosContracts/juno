package helpers

import (
	"context"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

func RegisterFeeShare(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, contract, withdrawAddr string) {
	// TF gas to create cost 2mil, so we set to 2.5 to be safe
	cmd := []string{
		"junod", "tx", "feeshare", "register", contract, user.FormattedAddress(), withdrawAddr,
		"--home", chain.HomeDir(),
		"--node", chain.GetRPCAddress(),
		"--chain-id", chain.Config().ChainID,
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	}
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	debugOutput(t, string(stdout))

	err = testutil.WaitForBlocks(ctx, 2, chain)
	require.NoError(t, err)
}
