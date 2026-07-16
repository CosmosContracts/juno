package suite

import (
	"encoding/json"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/cosmos/interchaintest/v10/testutil"
)

// RegisterCwHooksStaking registers a contract for staking hooks
func (s *E2ETestSuite) RegisterCwHooksStaking(chain *cosmos.CosmosChain, user ibc.Wallet, contractAddr string) {
	fees := sdk.NewCoins(sdk.NewCoin(chain.Config().Denom, math.NewInt(1000000)))
	s.cwHooksCmd(chain, "register", "staking", user, contractAddr, fees)
}

func (s *E2ETestSuite) RegisterCwHooksGovernance(chain *cosmos.CosmosChain, user ibc.Wallet, contractAddr string) {
	fees := sdk.NewCoins(sdk.NewCoin(chain.Config().Denom, math.NewInt(1000000)))
	s.cwHooksCmd(chain, "register", "gov", user, contractAddr, fees)
}

// UnregisterCwHooksStaking unregisters a contract from staking hooks
func (s *E2ETestSuite) UnregisterCwHooksStaking(chain *cosmos.CosmosChain, user ibc.Wallet, contractAddr string) {
	fees := sdk.NewCoins(sdk.NewCoin(chain.Config().Denom, math.NewInt(1000000)))
	s.cwHooksCmd(chain, "unregister", "staking", user, contractAddr, fees)
}

func (s *E2ETestSuite) UnregisterCwHooksGovernance(chain *cosmos.CosmosChain, user ibc.Wallet, contractAddr string) {
	fees := sdk.NewCoins(sdk.NewCoin(chain.Config().Denom, math.NewInt(1000000)))
	s.cwHooksCmd(chain, "unregister", "gov", user, contractAddr, fees)
}

// GetCwHooksStakingContracts retrieves all registered staking hook contracts
func (s *E2ETestSuite) GetCwHooksStakingContracts() []string {
	return s.getContracts(s.Chain, "staking")
}

func (s *E2ETestSuite) GetCwHooksGovernanceContracts() []string {
	return s.getContracts(s.Chain, "gov")
}

// GetCwStakingHookLastDelegationChange gets the last delegation change from a contract
func (s *E2ETestSuite) GetCwStakingHookLastDelegationChange(chain *cosmos.CosmosChain, contract string, _ string) GetCwHooksDelegationResponse {
	requireT := s.Require()
	var res GetCwHooksDelegationResponse
	err := s.SmartQueryString(chain, contract, `{"last_delegation_change":{}}`, &res)
	requireT.NoError(err)
	return res
}

// helpers
func (s *E2ETestSuite) cwHooksCmd(chain *cosmos.CosmosChain, command, module string, user ibc.Wallet, contractAddr string, fees sdk.Coins) {
	t := s.T()
	require := s.Require()

	stdout, err := s.ExecTx(
		s.Chain,
		user.KeyName(),
		false,
		false,
		"cw-hooks",
		command,
		user.FormattedAddress(),
		module,
		contractAddr,
		"--fees",
		fees.String(),
		"--gas",
		"auto",
	)
	require.NoError(err, "failed to execute cw-hooks command")

	s.DebugOutput(string(stdout))

	if err := testutil.WaitForBlocks(s.Ctx, 1, chain); err != nil {
		t.Fatal(err)
	}
}

func (s *E2ETestSuite) getContracts(chain *cosmos.CosmosChain, module string) []string {
	t := s.T()
	require := s.Require()
	cmd := []string{
		"junod", "query", "cw-hooks", "contracts", module,
		"--output", "json",
		"--node", chain.GetRPCAddress(),
	}

	stdout, _, err := chain.Exec(s.Ctx, cmd, nil)
	require.NoError(err)

	s.DebugOutput(string(stdout))

	type contractInfo struct {
		ContractAddress string `json:"contract_address"`
	}
	type contracts struct {
		Contracts []contractInfo `json:"contracts"`
	}

	var c contracts
	if err := json.Unmarshal(stdout, &c); err != nil {
		t.Fatal(err)
	}

	addrs := make([]string, 0, len(c.Contracts))
	for _, info := range c.Contracts {
		addrs = append(addrs, info.ContractAddress)
	}

	return addrs
}
