package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
)

// Register the clock contract
func RegisterClockContract(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, contract string) (*sdk.TxResponse, error) {
	cmd := []string{
		"clock", "register", user.FormattedAddress(), contract,
		"--home", chain.HomeDir(),
		"--fees", "500ujuno",
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	}
	node := chain.GetNode()
	txHash, err := node.ExecTx(ctx, user.KeyName(), cmd...)
	if err != nil {
		t.Fatal(err)
	}
	// convert stdout into a TxResponse
	txRes, err := chain.GetTransaction(txHash)
	if err != nil {
		t.Fatal(err)
	}

	err = testutil.WaitForBlocks(ctx, 2, chain)
	require.NoError(t, err)

	return txRes, nil
}

// Unregister the clock contract
func UnregisterClockContract(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, contract string) (*sdk.TxResponse, error) {
	cmd := []string{
		"clock", "unregister", user.FormattedAddress(), contract,
		"--home", chain.HomeDir(),
		"--fees", "500ujuno",
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	}
	node := chain.GetNode()
	txHash, err := node.ExecTx(ctx, user.KeyName(), cmd...)
	if err != nil {
		t.Fatal(err)
	}
	// convert stdout into a TxResponse
	txRes, err := chain.GetTransaction(txHash)
	if err != nil {
		t.Fatal(err)
	}

	err = testutil.WaitForBlocks(ctx, 2, chain)
	require.NoError(t, err)

	return txRes, nil
}

// Unjail the clock contract
func UnjailClockContract(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, contract string) (*sdk.TxResponse, error) {
	cmd := []string{
		"clock", "unjail", user.FormattedAddress(), contract,
		"--home", chain.HomeDir(),
		"--fees", "500ujuno",
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	}
	node := chain.GetNode()
	txHash, err := node.ExecTx(ctx, user.KeyName(), cmd...)
	if err != nil {
		t.Fatal(err)
	}
	// convert stdout into a TxResponse
	txRes, err := chain.GetTransaction(txHash)
	if err != nil {
		t.Fatal(err)
	}

	err = testutil.WaitForBlocks(ctx, 2, chain)
	require.NoError(t, err)

	return txRes, nil
}

type ClockContract struct {
	ClockContract struct {
		ContractAddress string `json:"contract_address"`
		IsJailed        bool   `json:"is_jailed"`
	} `json:"clock_contract"`
}

// Get the clock contract
func GetClockContract(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, contract string) ClockContract {
	var res ClockContract

	cmd := getClockQueryCommand(contract, chain)
	node := chain.GetNode()
	stdout, _, err := node.ExecQuery(ctx, cmd...)
	require.NoError(t, err)

	fmt.Println(string(stdout))

	if err := json.Unmarshal(stdout, &res); err != nil {
		t.Fatal(err)
	}

	return res
}

// Validate a contract is not registered with the clock module
func ValidateNoClockContract(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, contract string) {
	cmd := getClockQueryCommand(contract, chain)
	_, _, err := chain.Exec(ctx, cmd, nil)
	require.Error(t, err)
}

// Get the clock query command
func getClockQueryCommand(contract string, chain *cosmos.CosmosChain) []string {
	return []string{
		"junod", "query", "clock", "contract", contract,
		"--output", "json",
		"--node", chain.GetRPCAddress(),
	}
}
