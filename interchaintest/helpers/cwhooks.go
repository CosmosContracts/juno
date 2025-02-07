package helpers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
)

// Register
func RegisterCwHooksStaking(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, contractAddr string) {
	cwHooksCmd(t, ctx, chain, user, "register", "staking", contractAddr)
}

func RegisterCwHooksGovernance(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, contractAddr string) {
	cwHooksCmd(t, ctx, chain, user, "register", "governance", contractAddr)
}

// UnRegister
func UnregisterCwHooksStaking(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, contractAddr string) {
	cwHooksCmd(t, ctx, chain, user, "unregister", "staking", contractAddr)
}

func UnregisterCwHooksGovernance(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, contractAddr string) {
	cwHooksCmd(t, ctx, chain, user, "unregister", "governance", contractAddr)
}

// Get Contracts
func GetCwHooksStakingContracts(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain) []string {
	return getContracts(t, ctx, chain, "staking-contracts")
}

func GetCwHooksGovernanceContracts(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain) []string {
	return getContracts(t, ctx, chain, "governance-contracts")
}

// Contract specific
func GetCwStakingHookLastDelegationChange(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, contract string, uaddr string) GetCwHooksDelegationResponse {
	var res GetCwHooksDelegationResponse
	err := SmartQueryString(t, ctx, chain, contract, `{"last_delegation_change":{}}`, &res)
	require.NoError(t, err)
	return res
}

// helpers
func cwHooksCmd(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, command, module, contractAddr string) {
	cmd := []string{
		"junod", "tx", "cw-hooks", command, module, contractAddr,
		"--home", chain.HomeDir(),
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

func getContracts(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, subCmd string) []string {
	cmd := []string{
		"junod", "query", "cw-hooks", subCmd,
		"--output", "json",
	}

	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	debugOutput(t, string(stdout))

	type contracts struct {
		Contracts []string `json:"contracts"`
	}

	var c contracts
	if err := json.Unmarshal(stdout, &c); err != nil {
		t.Fatal(err)
	}

	return c.Contracts
}
