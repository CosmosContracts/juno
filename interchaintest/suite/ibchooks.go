package suite

import (
	"strings"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/stretchr/testify/require"
)

func (s *E2ETestSuite) GetIBCHooksUserAddress(chain *cosmos.CosmosChain, channel, uaddr string) string {
	t := s.T()
	// junod q ibchooks wasm-sender channel-0 "juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl"
	cmd := []string{
		"junod", "query", "ibchooks", "wasm-sender", channel, uaddr,
		"--output", "json",
		"--node", chain.GetRPCAddress(),
	}

	// This query does not return a type, just prints the string.
	stdout, _, err := chain.Exec(s.Ctx, cmd, nil)
	require.NoError(t, err)

	address := strings.Replace(string(stdout), "\n", "", -1)
	return address
}

func (s *E2ETestSuite) GetIBCHookTotalFunds(chain *cosmos.CosmosChain, contract string, uaddr string) GetTotalFundsResponse {
	require := s.Require()
	var res GetTotalFundsResponse
	err := chain.QueryContract(s.Ctx, contract, ContractQueryMsg{GetTotalFunds: &GetTotalFundsQuery{Addr: uaddr}}, &res)
	require.NoError(err)
	return res
}

func (s *E2ETestSuite) GetIBCHookCount(chain *cosmos.CosmosChain, contract string, uaddr string) GetCountResponse {
	require := s.Require()
	var res GetCountResponse
	err := chain.QueryContract(s.Ctx, contract, ContractQueryMsg{GetCount: &GetCountQuery{Addr: uaddr}}, &res)
	require.NoError(err)
	return res
}
