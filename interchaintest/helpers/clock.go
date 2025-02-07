package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
)

// Register the clock contract
func RegisterClockContract(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, contract string) {
	cmd := []string{
		"junod", "tx", "clock", "register", contract,
		"--home", chain.HomeDir(),
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

// Unregister the clock contract
func UnregisterClockContract(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, contract string) {
	cmd := []string{
		"junod", "tx", "clock", "unregister", contract,
		"--home", chain.HomeDir(),
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

// Unjail the clock contract
func UnjailClockContract(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, contract string) {
	cmd := []string{
		"junod", "tx", "clock", "unjail", contract,
		"--home", chain.HomeDir(),
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

type ClockContract struct {
	ClockContract struct {
		ContractAddress string `json:"contract_address"`
		IsJailed        bool   `json:"is_jailed"`
	} `json:"clock_contract"`
}

// Get the clock contract
func GetClockContract(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, contract string) ClockContract {
	var res ClockContract

	cmd := getClockQueryCommand(contract)
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	fmt.Println(string(stdout))

	if err := json.Unmarshal(stdout, &res); err != nil {
		t.Fatal(err)
	}

	return res
}

// Validate a contract is not registered with the clock module
func ValidateNoClockContract(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, contract string) {
	cmd := getClockQueryCommand(contract)
	_, _, err := chain.Exec(ctx, cmd, nil)
	require.Error(t, err)
}

// Get the clock query command
func getClockQueryCommand(contract string) []string {
	return []string{
		"junod", "query", "clock", "contract", contract,
		"--output", "json",
	}
}
