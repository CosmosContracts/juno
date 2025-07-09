package suite

import (
	"encoding/json"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
)

// RegisterCwHooksStaking registers a contract for staking hooks
func (s *E2ETestSuite) RegisterCwHooksStaking(chain *cosmos.CosmosChain, user ibc.Wallet, contractAddr string) {
	fees := sdk.NewCoins(sdk.NewCoin(chain.Config().Denom, math.NewInt(1000000)))
	s.cwHooksCmd(chain, "register-staking", user, contractAddr, fees)
}

func (s *E2ETestSuite) RegisterCwHooksGovernance(chain *cosmos.CosmosChain, user ibc.Wallet, contractAddr string) {
	fees := sdk.NewCoins(sdk.NewCoin(chain.Config().Denom, math.NewInt(1000000)))
	s.cwHooksCmd(chain, "register-governance", user, contractAddr, fees)
}

// UnregisterCwHooksStaking unregisters a contract from staking hooks
func (s *E2ETestSuite) UnregisterCwHooksStaking(chain *cosmos.CosmosChain, user ibc.Wallet, contractAddr string) {
	fees := sdk.NewCoins(sdk.NewCoin(chain.Config().Denom, math.NewInt(1000000)))
	s.cwHooksCmd(chain, "unregister-staking", user, contractAddr, fees)
}

func (s *E2ETestSuite) UnregisterCwHooksGovernance(chain *cosmos.CosmosChain, user ibc.Wallet, contractAddr string) {
	fees := sdk.NewCoins(sdk.NewCoin(chain.Config().Denom, math.NewInt(1000000)))
	s.cwHooksCmd(chain, "unregister-governance", user, contractAddr, fees)
}

// GetCwHooksStakingContracts retrieves all registered staking hook contracts
func (s *E2ETestSuite) GetCwHooksStakingContracts() []string {
	return s.getContracts(s.Chain, "staking-contracts")
}

func (s *E2ETestSuite) GetCwHooksGovernanceContracts() []string {
	return s.getContracts(s.Chain, "governance-contracts")
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
func (s *E2ETestSuite) cwHooksCmd(chain *cosmos.CosmosChain, command string, user ibc.Wallet, contractAddr string, fees sdk.Coins) {
	t := s.T()
	require := s.Require()

	stdout, err := s.ExecTx(
		s.Chain,
		user.KeyName(),
		false,
		false,
		"cw-hooks",
		command,
		contractAddr,
		user.FormattedAddress(),
		"--fees",
		fees.String(),
		"--gas",
		"auto",
	)
	require.NoError(err, "failed to execute cw-hooks command")

	s.DebugOutput(string(stdout))

	if err := testutil.WaitForBlocks(s.Ctx, 2, chain); err != nil {
		t.Fatal(err)
	}
}

func (s *E2ETestSuite) getContracts(chain *cosmos.CosmosChain, subCmd string) []string {
	t := s.T()
	require := s.Require()
	cmd := []string{
		"junod", "query", "cw-hooks", subCmd,
		"--output", "json",
		"--node", chain.GetRPCAddress(),
	}

	stdout, _, err := chain.Exec(s.Ctx, cmd, nil)
	require.NoError(err)

	s.DebugOutput(string(stdout))

	type contracts struct {
		Contracts []string `json:"contracts"`
	}

	var c contracts
	if err := json.Unmarshal(stdout, &c); err != nil {
		t.Fatal(err)
	}

	return c.Contracts
}
