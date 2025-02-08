package helpers

import (
	"context"
	"strings"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/stretchr/testify/require"
)

func GetIBCHooksUserAddress(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, channel, uaddr string) string {
	// junod q ibchooks wasm-sender channel-0 "juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl"
	cmd := []string{
		"junod", "query", "ibchooks", "wasm-sender", channel, uaddr,
		"--output", "json",
		"--node", chain.GetRPCAddress(),
	}

	// This query does not return a type, just prints the string.
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	address := strings.Replace(string(stdout), "\n", "", -1)
	return address
}

func GetIBCHookTotalFunds(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, contract string, uaddr string) GetTotalFundsResponse {
	var res GetTotalFundsResponse
	err := chain.QueryContract(ctx, contract, QueryMsg{GetTotalFunds: &GetTotalFundsQuery{Addr: uaddr}}, &res)
	require.NoError(t, err)
	return res
}

func GetIBCHookCount(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, contract string, uaddr string) GetCountResponse {
	var res GetCountResponse
	err := chain.QueryContract(ctx, contract, QueryMsg{GetCount: &GetCountQuery{Addr: uaddr}}, &res)
	require.NoError(t, err)
	return res
}
