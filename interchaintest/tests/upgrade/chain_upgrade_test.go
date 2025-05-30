package upgrade_test

import (
	"strconv"
	"testing"

	"cosmossdk.io/math"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"
)

const (
	chainName   = "juno"
	upgradeName = "v30"
)

// baseChain is the current version of the chain that will be upgraded from
var baseChain = ibc.DockerImage{
	Repository: e2esuite.JunoRepo,
	Version:    "v29.0.0",
	UIDGID:     "1025:1025",
}

type BasicUpgradeTestSuite struct {
	*e2esuite.E2ETestSuite
}

func TestBasicUpgradeTestSuite(t *testing.T) {
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

	cfg := e2esuite.DefaultConfig
	cfg.ModifyGenesis = cosmos.ModifyGenesis(previousVersionGenesis)
	cfg.Images = []ibc.DockerImage{baseChain}

	numValidators := 4
	numFullNodes := 1

	spec := &interchaintest.ChainSpec{
		ChainName:     "juno-upgrade",
		Name:          "juno-upgrade",
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

	t.Parallel()
	t.Cleanup(func() {
		_ = s.Ic.Close()
	})

	testSuite := &BasicUpgradeTestSuite{E2ETestSuite: s}
	suite.Run(t, testSuite)
}

func (s *BasicUpgradeTestSuite) TestBasicChainUpgrade() {
	t := s.T()
	require := s.Require()
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	fees := sdk.NewCoins(sdk.NewCoin(s.Denom, math.NewInt(100_000)))
	repo, version := e2esuite.GetDockerImageInfo()
	user := s.GetAndFundTestUser(t.Name(), 10_000_000_000, s.Chain)

	// execute a contract before the upgrade
	beforeContract := s.StdExecute(s.Chain, user)

	// upgrade
	height, err := s.Chain.Height(s.Ctx)
	require.NoError(err, "error fetching height before submit upgrade proposal")

	haltHeight := height + e2esuite.DefaultHaltHeightDelta
	proposalID := s.SubmitSoftwareUpgradeProposal(s.Chain, user, upgradeName, haltHeight, e2esuite.DefaultAuthority)

	proposalIDInt, err := strconv.ParseUint(proposalID, 10, 64)
	require.NoError(err, "failed to parse proposal ID")

	s.ValidatorVoting(s.Chain, proposalIDInt, height, haltHeight)
	s.UpgradeNodes(s.Chain, s.DockerClient, haltHeight, repo, version)

	// confirm we can execute against the beforeContract (ref: v20 upgrade patch)
	_, err = s.ExecuteMsgWithFeeReturn(s.Chain, user, beforeContract, "", `{"increment":{}}`, false, fees)
	require.NoError(err)

	// Post Upgrade: Conformance Validation
	s.ConformanceCosmWasm(s.Chain, user)
}
