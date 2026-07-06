package upgrade_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/cosmos/interchaintest/v10/testutil"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"
)

const (
	upgradeName = "v30"
	// Deliberately different from the handler's 25M fallback so the
	// post-upgrade assertion proves feemarket read consensus max_gas
	// rather than silently falling back.
	expectedConsensusMaxGas = uint64(30_000_000)
)

// baseChain is the current version of the chain that will be upgraded from
var baseChain = ibc.DockerImage{
	Repository: e2esuite.JunoRepo,
	Version:    "v29.0.0",
	UIDGID:     "1025:1025",
}

type UpgradeTestSuite struct {
	*e2esuite.E2ETestSuite
}

func TestUpgradeTestSuite(t *testing.T) {
	cfg := e2esuite.DefaultConfig
	cfg.Images = []ibc.DockerImage{baseChain}

	numValidators := 2
	numFullNodes := 1

	previousVersionGenesis := []cosmos.GenesisKV{
		{
			Key:   "app_state.gov.params.voting_period",
			Value: e2esuite.DefaultVotingPeriod,
		},
		{
			Key:   "app_state.gov.params.max_deposit_period",
			Value: e2esuite.DefaultMaxDepositPeriod,
		},
		{
			Key:   "app_state.gov.params.min_deposit.0.denom",
			Value: e2esuite.DefaultDenom,
		},
		{
			Key:   "consensus.params.block.max_gas",
			Value: strconv.FormatUint(expectedConsensusMaxGas, 10),
		},
	}
	cfg.ModifyGenesis = cosmos.ModifyGenesis(previousVersionGenesis)

	spec := &interchaintest.ChainSpec{
		ChainName:     "juno",
		Name:          "juno",
		NumValidators: &numValidators,
		NumFullNodes:  &numFullNodes,
		Version:       baseChain.Version,
		NoHostMount:   &e2esuite.DefaultNoHostMount,
		ChainConfig:   cfg,
	}
	specs := []*interchaintest.ChainSpec{spec}

	s := e2esuite.NewE2ETestSuite(
		specs,
		e2esuite.DefaultTxCfg,
	)

	t.Cleanup(func() {
		_ = s.Ic.Close()
	})

	testSuite := &UpgradeTestSuite{E2ETestSuite: s}
	suite.Run(t, testSuite)
}

func (s *UpgradeTestSuite) TestV30ChainUpgrade() {
	t := s.T()
	require := s.Require()
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	fees := sdk.NewCoins(sdk.NewCoin(s.Denom, math.NewInt(100_000)))
	user := s.GetAndFundTestUser(t.Name(), 10_000_000_000, s.Chain)

	// prepare a cw-hooks staking contract and ensure it is functional prior to upgrade
	const cwHooksExampleWasm = "../../contracts/juno_staking_hooks_example.wasm"
	_, hookContract := s.SetupContract(s.Chain, user.KeyName(), cwHooksExampleWasm, `{}`, false, fees)
	s.legacyCwHooksCmd("register-staking", user, hookContract, fees)
	stakingContracts := s.legacyGetCwHooksContracts("staking-contracts")
	require.Contains(stakingContracts, hookContract, "cw-hooks contract was not registered with the staking module")

	vals := s.QueryValidators(s.Chain)
	require.NotEmpty(vals, "expected at least one validator")
	valoper := vals[0]
	initialStakeAmt := int64(1_000_000)
	initialStakeCoins := fmt.Sprintf("%d%s", initialStakeAmt, s.Denom)
	s.StakeTokens(s.Chain, user, valoper.String(), initialStakeCoins, fees, false)

	initialHookState := s.GetCwStakingHookLastDelegationChange(s.Chain, hookContract, user.FormattedAddress())
	require.NotNil(initialHookState.Data, "pre-upgrade cw-hooks contract did not record the delegation event")
	require.Equal(user.FormattedAddress(), initialHookState.Data.DelegatorAddress)
	require.Equal(valoper, sdk.MustValAddressFromBech32(initialHookState.Data.ValidatorAddress))
	require.Equal(fmt.Sprintf("%d.000000000000000000", initialStakeAmt), initialHookState.Data.Shares)

	// upgrade
	height, err := s.Chain.Height(s.Ctx)
	require.NoError(err, "error fetching height before submit upgrade proposal")

	haltHeight := height + e2esuite.DefaultHaltHeightDelta
	proposalID := s.SubmitSoftwareUpgradeProposal(s.Chain, user, upgradeName, haltHeight, e2esuite.DefaultAuthority)

	proposalIDInt, err := strconv.ParseUint(proposalID, 10, 64)
	require.NoError(err, "failed to parse proposal ID")

	s.ValidatorVoting(s.Chain, proposalIDInt, height, haltHeight)
	repo, version := e2esuite.GetDockerImageInfo()
	s.UpgradeNodes(s.Chain, s.DockerClient, haltHeight, repo, version)

	// verify cw-hooks state survived the migration
	postUpgradeContracts := s.GetCwHooksStakingContracts()
	require.Contains(postUpgradeContracts, hookContract, "cw-hooks contract no longer registered after migration")

	cwHooksParams := s.QueryCwHooksParams()
	require.Equal(uint64(3), cwHooksParams.ContractFailureRemovalThreshold,
		"cw-hooks contract failure removal threshold should be migrated")

	additionalStakeAmt := int64(500_000)
	additionalStakeCoins := fmt.Sprintf("%d%s", additionalStakeAmt, s.Denom)
	s.StakeTokens(s.Chain, user, valoper.String(), additionalStakeCoins, fees, false)

	postHookState := s.GetCwStakingHookLastDelegationChange(s.Chain, hookContract, user.FormattedAddress())
	require.NotNil(postHookState.Data, "post-upgrade cw-hooks contract failed to record delegation event")
	require.Equal(user.FormattedAddress(), postHookState.Data.DelegatorAddress)
	require.Equal(valoper, sdk.MustValAddressFromBech32(postHookState.Data.ValidatorAddress))
	require.Equal(fmt.Sprintf("%d.000000000000000000", initialStakeAmt+additionalStakeAmt), postHookState.Data.Shares)

	// --- new v30 modules: feemarket (added store) must be initialized ---
	// Params/State/GasPrice must all resolve to sane, positive dynamic-fee
	// values after InitGenesis of the freshly-added feemarket store.
	feemarketParams := s.QueryFeemarketParams()
	require.True(feemarketParams.Enabled, "feemarket should be enabled after upgrade")
	require.False(feemarketParams.MinBaseGasPrice.IsNil(), "feemarket min base gas price should be set")
	require.True(feemarketParams.MinBaseGasPrice.IsPositive(), "feemarket min base gas price should be positive")
	require.Equal(expectedConsensusMaxGas, feemarketParams.MaxBlockUtilization,
		"feemarket max block utilization should match consensus block max gas")

	feemarketState := s.QueryFeemarketState()
	require.False(feemarketState.BaseGasPrice.IsNil(), "feemarket base gas price state should be set")
	require.True(feemarketState.BaseGasPrice.IsPositive(), "feemarket base gas price should be positive after upgrade")

	gasPrice := s.QueryFeemarketGasPrice(s.Denom)
	require.Equal(s.Denom, gasPrice.Denom)
	require.True(gasPrice.Amount.IsPositive(), "feemarket gas price for %s should be positive", s.Denom)

	// --- new v30 modules: voting-snapshot (added store) must be initialized ---
	// InitGenesis seeds active delegators, and the post-upgrade delegation
	// above writes a fresh snapshot. Both the module params and the gRPC power
	// queries must return sane values.
	vsParams := s.QueryVotingSnapshotParams()
	require.Positive(vsParams.PruneInterval, "voting-snapshot prune interval should be a sane positive default")

	snapHeight, err := s.Chain.Height(s.Ctx)
	require.NoError(err, "error fetching height for voting-snapshot query")

	// Pre-upgrade delegator's snapshotted (LST-excluded) voting power must be
	// positive — proving the store was initialized and hooks/backfill ran.
	powerStr := s.QueryVotingPowerAt(user.FormattedAddress(), snapHeight)
	power, ok := math.NewIntFromString(powerStr)
	require.True(ok, "voting power %q should parse as an integer", powerStr)
	require.True(power.IsPositive(), "pre-upgrade delegator should have positive snapshotted voting power")

	// It should not exceed the user's actual bonded delegation.
	delegation := s.QueryStakingDelegation(user.FormattedAddress(), valoper.String())
	require.True(power.LTE(delegation.Balance.Amount),
		"snapshotted voting power (%s) should not exceed bonded delegation (%s)", power, delegation.Balance.Amount)

	// Chain-wide total voting power must be at least this single delegator's.
	totalStr := s.QueryTotalVotingPowerAt(snapHeight)
	total, ok := math.NewIntFromString(totalStr)
	require.True(ok, "total voting power %q should parse as an integer", totalStr)
	require.True(total.GTE(power), "total voting power (%s) should be >= delegator power (%s)", total, power)
}

func (s *UpgradeTestSuite) legacyCwHooksCmd(command string, user ibc.Wallet, contractAddr string, fees sdk.Coins) {
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
	require.NoError(err, "failed to execute legacy cw-hooks command")

	s.DebugOutput(string(stdout))

	if err := testutil.WaitForBlocks(s.Ctx, 2, s.Chain); err != nil {
		t.Fatal(err)
	}
}

func (s *UpgradeTestSuite) legacyGetCwHooksContracts(subCmd string) []string {
	t := s.T()
	require := s.Require()
	cmd := []string{
		"junod", "query", "cw-hooks", subCmd,
		"--output", "json",
		"--node", s.Chain.GetRPCAddress(),
	}

	stdout, _, err := s.Chain.Exec(s.Ctx, cmd, nil)
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
