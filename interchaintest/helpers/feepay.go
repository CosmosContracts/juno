package helpers

import (
	"context"
	"fmt"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

func RegisterFeePay(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, contract string, walletLimit int) {
	cmd := []string{
		"junod", "tx", "feepay", "register", contract, fmt.Sprintf("%d", walletLimit),
		"--node", chain.GetRPCAddress(),
		"--home", chain.HomeDir(),
		"--chain-id", chain.Config().ChainID,
		"--fees", "500ujuno",
		"--from", user.KeyName(),
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

func FundFeePayContract(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, contract string, amountCoin string) {
	cmd := []string{
		"junod", "tx", "feepay", "fund", contract, amountCoin,
		"--node", chain.GetRPCAddress(),
		"--home", chain.HomeDir(),
		"--chain-id", chain.Config().ChainID,
		"--fees", "500ujuno",
		"--from", user.KeyName(),
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

func UpdateFeePayWalletLimit(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, contract string, newLimit uint64) {
	cmd := []string{
		"junod", "tx", "feepay", "update-wallet-limit", contract, fmt.Sprintf("%d", newLimit),
		"--node", chain.GetRPCAddress(),
		"--home", chain.HomeDir(),
		"--chain-id", chain.Config().ChainID,
		"--from", user.KeyName(),
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
