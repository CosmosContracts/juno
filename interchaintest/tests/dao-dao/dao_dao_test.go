// Package dao_dao_test exercises DAO DAO v2.7.0 contracts against the
// v30 chain binary. Per planning/09-deferred-work.md §A1, this is the
// gate test for "DAO DAO contracts continue to work after the wasmvm
// v3 / sdk v0.53.7 / ibc-go v10 upgrade."
//
// Three legs:
//   1. cw4-group voting + proposal-single (smallest path; proves the
//      module contracts instantiate + interact)
//   2. cw20-staked voting + proposal-single (heavier path; cw20 token
//      + staking module contracts plus the voting+proposal pair)
//   3. wasmbinding smoke for VotingPowerAt (target of the new
//      x/voting-snapshot module)
//
// Run with `make ictest-dao-dao` once that target lands.
package daodao_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/stretchr/testify/suite"

	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"
)

type DaoDaoTestSuite struct {
	*e2esuite.E2ETestSuite
}

func TestDaoDaoTestSuite(t *testing.T) {
	s := e2esuite.NewE2ETestSuite(
		[]*interchaintest.ChainSpec{e2esuite.DefaultSpec},
		e2esuite.DefaultTxCfg,
	)

	t.Parallel()
	t.Cleanup(func() {
		if s.Ic != nil {
			_ = s.Ic.Close()
		}
	})

	testSuite := &DaoDaoTestSuite{E2ETestSuite: s}
	suite.Run(t, testSuite)
}

// TestCw4GroupDao instantiates a minimal DAO DAO setup with cw4-group
// voting + proposal-single, opens a no-op proposal, votes it through,
// and executes it. Pass criterion: every step returns success and the
// proposal moves through Open → Passed → Executed.
func (s *DaoDaoTestSuite) TestCw4GroupDao() {
	t := s.T()
	require := s.Require()

	user := s.GetAndFundTestUser(t.Name(), 10_000_000_000, s.Chain)
	fees := sdk.NewCoins(sdk.NewCoin(s.Denom, math.NewInt(100_000)))

	// Store the four contracts the cw4-group path needs:
	//   dao-dao-core, dao-proposal-single, dao-voting-cw4, cw4-group
	cw4GroupCodeID := s.StoreContract(s.Chain, user.KeyName(), "../../contracts/cw4_group.wasm", fees)
	votingCodeID := s.StoreContract(s.Chain, user.KeyName(), "../../contracts/dao_voting_cw4.wasm", fees)
	proposalCodeID := s.StoreContract(s.Chain, user.KeyName(), "../../contracts/dao_proposal_single.wasm", fees)
	coreCodeID := s.StoreContract(s.Chain, user.KeyName(), "../../contracts/dao_dao_core.wasm", fees)

	// Instantiate the DAO. The dao-dao-core constructor takes
	// instantiate-info structs for the voting and proposal modules;
	// see dao-contracts/packages/dao-interface for the schema.
	daoMsg := buildDaoInstantiate(user.FormattedAddress(), votingCodeID, proposalCodeID, cw4GroupCodeID)
	dao, err := s.InstantiateContract(s.Chain, user.KeyName(), coreCodeID, daoMsg, fees, false, false)
	require.NoError(err)
	require.NotEmpty(dao)

	// Discover the child voting + proposal contract addresses.
	voting := queryVotingModule(t, s.Chain, dao)
	proposal := queryProposalModule(t, s.Chain, dao)
	require.NotEmpty(voting)
	require.NotEmpty(proposal)

	// Open a no-op proposal, vote yes from the single member, execute.
	proposalID := openProposal(t, s.Chain, user.KeyName(), proposal, "test", "no-op proposal")
	voteOnProposal(t, s.Chain, user.KeyName(), proposal, proposalID, "yes")
	executeProposal(t, s.Chain, user.KeyName(), proposal, proposalID)

	status := queryProposalStatus(t, s.Chain, proposal, proposalID)
	require.Equal("executed", status)
}

// TestCw20StakedDao exercises the staked-token voting path. Stakers
// contribute voting power proportional to their staked balance; the
// proposal threshold is a percentage of the snapshot supply. Pass
// criterion: a single staker can pass a no-op proposal that crosses
// the threshold.
func (s *DaoDaoTestSuite) TestCw20StakedDao() {
	t := s.T()
	t.Skip("TODO(v30.x): implement once the cw4-group leg passes — same scaffold, swap voting module for cw20-staked + add cw20 token + cw20-stake setup")
}

// TestWasmbindingsVotingPowerAt verifies the x/voting-snapshot
// custom binding. Deploys a small "echo" contract that calls
// JunoQuery::VotingPowerAt and emits the result as an event;
// asserts the result matches the staker's bonded amount.
func (s *DaoDaoTestSuite) TestWasmbindingsVotingPowerAt() {
	t := s.T()
	t.Skip("TODO(v30.x): build a minimal Rust contract that invokes JunoQuery::VotingPowerAt; embed wasm at interchaintest/contracts/voting_power_probe.wasm")
}

// helpers — TODO(v30.x): flesh out once the test runs in CI and we can
// iterate on real msg shapes. Keeping these as stubs so the test file
// compiles; the cw4-group leg's first concrete pass is the next-session
// goal.

func buildDaoInstantiate(creator string, votingCodeID, proposalCodeID, cw4CodeID string) string {
	// Skeleton — fill in once we have a concrete schema reference.
	type instantiateInfo struct {
		CodeID  string          `json:"code_id"`
		Msg     json.RawMessage `json:"msg"`
		Funds   []sdk.Coin      `json:"funds"`
		Label   string          `json:"label"`
		Admin   *string         `json:"admin"`
	}
	_ = instantiateInfo{}
	_ = creator
	_ = votingCodeID
	_ = proposalCodeID
	_ = cw4CodeID
	return `{"name":"test-dao","description":"v30 ictest DAO","voting_module_instantiate_info":null,"proposal_modules_instantiate_info":[]}`
}

func queryVotingModule(t *testing.T, chain *cosmos.CosmosChain, dao string) string {
	_ = chain
	_ = dao
	t.Skip("queryVotingModule helper unimplemented")
	return ""
}

func queryProposalModule(t *testing.T, chain *cosmos.CosmosChain, dao string) string {
	_ = chain
	_ = dao
	t.Skip("queryProposalModule helper unimplemented")
	return ""
}

func openProposal(t *testing.T, chain *cosmos.CosmosChain, key, proposal, title, desc string) uint64 {
	_, _, _, _, _ = chain, key, proposal, title, desc
	t.Skip("openProposal helper unimplemented")
	return 0
}

func voteOnProposal(t *testing.T, chain *cosmos.CosmosChain, key, proposal string, id uint64, vote string) {
	_, _, _, _, _ = chain, key, proposal, id, vote
	t.Skip("voteOnProposal helper unimplemented")
}

func executeProposal(t *testing.T, chain *cosmos.CosmosChain, key, proposal string, id uint64) {
	_, _, _, _ = chain, key, proposal, id
	t.Skip("executeProposal helper unimplemented")
}

func queryProposalStatus(t *testing.T, chain *cosmos.CosmosChain, proposal string, id uint64) string {
	_ = context.Background()
	_, _, _ = chain, proposal, id
	t.Skip("queryProposalStatus helper unimplemented")
	return fmt.Sprintf("status-stub-%d", id)
}
