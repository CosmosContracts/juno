package upgrade_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/ibc"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"
)

const (
	upgradeName = "v31"
	// libwasmvm carrying the public Wasmer security fix. Every upgraded node
	// must report exactly this version or it is still exposed.
	expectedLibwasmvmVersion = "3.0.8"
)

// baseChain is the current version of the chain that will be upgraded from
var baseChain = ibc.DockerImage{
	Repository: e2esuite.JunoRepo,
	Version:    "v30.0.0",
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
		if s.Ic != nil {
			_ = s.Ic.Close()
		}
	})

	testSuite := &UpgradeTestSuite{E2ETestSuite: s}
	suite.Run(t, testSuite)
}

func (s *UpgradeTestSuite) TestV31ChainUpgrade() {
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
	s.RegisterCwHooksStaking(s.Chain, user, hookContract)
	require.Contains(s.GetCwHooksStakingContracts(), hookContract, "cw-hooks contract was not registered with the staking module")

	vals := s.QueryValidators(s.Chain)
	require.NotEmpty(vals, "expected at least one validator")
	valoper := vals[0]
	initialStakeAmt := int64(1_000_000)
	initialStakeCoins := fmt.Sprintf("%d%s", initialStakeAmt, s.Denom)
	s.StakeTokens(s.Chain, user, valoper.String(), initialStakeCoins, fees, false)

	initialHookState := s.GetCwStakingHookLastDelegationChange(s.Chain, hookContract, user.FormattedAddress())
	require.NotNil(initialHookState.Data, "pre-upgrade cw-hooks contract did not record the delegation event")
	require.Equal(fmt.Sprintf("%d.000000000000000000", initialStakeAmt), initialHookState.Data.Shares)

	// v31 carries no intentional state transformation: capture params that
	// must survive the upgrade unchanged.
	feemarketParamsBefore := s.QueryFeemarketParams()
	cwHooksParamsBefore := s.QueryCwHooksParams()
	votingSnapshotParamsBefore := s.QueryVotingSnapshotParams()

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

	// every node must run the patched VM
	for _, node := range s.Chain.Nodes() {
		stdout, _, err := node.ExecBin(s.Ctx, "query", "wasm", "libwasmvm-version")
		require.NoError(err, "failed to query libwasmvm version on %s", node.Name())
		require.Equal(expectedLibwasmvmVersion, strings.TrimSpace(string(stdout)),
			"node %s runs an unexpected libwasmvm version", node.Name())
	}

	// params must be unchanged by the v31 handler
	require.Equal(feemarketParamsBefore, s.QueryFeemarketParams(), "feemarket params changed across v31")
	require.Equal(cwHooksParamsBefore, s.QueryCwHooksParams(), "cw-hooks params changed across v31")
	require.Equal(votingSnapshotParamsBefore, s.QueryVotingSnapshotParams(), "voting-snapshot params changed across v31")

	// feemarket must still price fees
	gasPrice := s.QueryFeemarketGasPrice(s.Denom)
	require.Equal(s.Denom, gasPrice.Denom)
	require.True(gasPrice.Amount.IsPositive(), "feemarket gas price for %s should be positive", s.Denom)

	// the pre-upgrade contract must still execute under the new VM: the
	// staking hook is a sudo call into the contract stored before the upgrade
	require.Contains(s.GetCwHooksStakingContracts(), hookContract, "cw-hooks contract no longer registered after upgrade")

	additionalStakeAmt := int64(500_000)
	additionalStakeCoins := fmt.Sprintf("%d%s", additionalStakeAmt, s.Denom)
	s.StakeTokens(s.Chain, user, valoper.String(), additionalStakeCoins, fees, false)

	postHookState := s.GetCwStakingHookLastDelegationChange(s.Chain, hookContract, user.FormattedAddress())
	require.NotNil(postHookState.Data, "post-upgrade cw-hooks contract failed to record delegation event")
	require.Equal(user.FormattedAddress(), postHookState.Data.DelegatorAddress)
	require.Equal(valoper, sdk.MustValAddressFromBech32(postHookState.Data.ValidatorAddress))
	require.Equal(fmt.Sprintf("%d.000000000000000000", initialStakeAmt+additionalStakeAmt), postHookState.Data.Shares)

	// a contract stored after the upgrade must also instantiate and run
	_, postContract := s.SetupContract(s.Chain, user.KeyName(), cwHooksExampleWasm, `{}`, false, fees)
	require.NotEmpty(postContract, "failed to instantiate a contract after upgrade")
}
