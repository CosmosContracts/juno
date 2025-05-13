package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	feepaytypes "github.com/CosmosContracts/juno/v29/x/feepay/types"
)

func RegisterFeePay(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, contract string, walletLimit int) {
	feePayContract := &feepaytypes.FeePayContract{
		ContractAddress: contract,
		Balance:         0,
		WalletLimit:     uint64(walletLimit),
	}

	metadataJSON, err := json.Marshal(feePayContract)
	require.NoError(t, err)

	cmd := []string{
		"junod", "tx", "feepay", "register", user.FormattedAddress(), string(metadataJSON),
		"--home", chain.HomeDir(),
		"--node", chain.GetRPCAddress(),
		"--chain-id", chain.Config().ChainID,
		"--fees", "500ujuno",
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
		"junod", "tx", "feepay", "fund", user.FormattedAddress(), contract, amountCoin,
		"--home", chain.HomeDir(),
		"--node", chain.GetRPCAddress(),
		"--chain-id", chain.Config().ChainID,
		"--fees", "500ujuno",
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
		"junod", "tx", "feepay", "update-wallet-limit", user.FormattedAddress(), contract, fmt.Sprintf("%d", newLimit),
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
