package fixes_test

import (
	"strconv"
	"testing"

	govv1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"

	buildertypes "github.com/skip-mev/pob/x/builder/types"
)

const (
	v28UpgradeName = "v28"
	v29UpgradeName = "v29"
)

var haltHeightDelta = int64(9) // will propose upgrade this many blocks in the future

// v27Chain is the first version of juno that will be upgraded from
var v27Chain = ibc.DockerImage{
	Repository: e2esuite.JunoRepo,
	Version:    "v27.0.0",
	UIDGID:     "1025:1025",
}

// v28Chain is the second version of juno that will be upgraded from
var v28Chain = ibc.DockerImage{
	Repository: e2esuite.JunoRepo,
	Version:    "v28.0.2",
	UIDGID:     "1025:1025",
}

// v29Chain is the final version of juno that will be upgraded to in this test
var v29Chain = ibc.DockerImage{
	Repository: e2esuite.JunoRepo,
	Version:    "v29.0.0",
	UIDGID:     "1025:1025",
}

type FixTestSuite struct {
	*e2esuite.E2ETestSuite
}

func TestFixTestSuite(t *testing.T) {
	v27VersionGenesis := []cosmos.GenesisKV{
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
	numVals, numNodes := 4, 0
	noHostMount := false
	cfg := e2esuite.DefaultConfig

	cfg.ModifyGenesis = cosmos.ModifyGenesis(v27VersionGenesis)
	cfg.Images = []ibc.DockerImage{v27Chain, v28Chain, v29Chain}
	buildertypes.RegisterInterfaces(cfg.EncodingConfig.InterfaceRegistry)

	spec := &interchaintest.ChainSpec{
		ChainName:     "juno",
		Name:          "juno",
		NumValidators: &numVals,
		NumFullNodes:  &numNodes,
		Version:       "v27.0.0",
		NoHostMount:   &noHostMount,
		ChainConfig:   cfg,
	}

	s := e2esuite.NewE2ETestSuite(
		[]*interchaintest.ChainSpec{spec},
		e2esuite.DefaultTxCfg,
	)

	t.Parallel()
	t.Cleanup(func() {
		_ = s.Ic.Close()
	})

	testSuite := &FixTestSuite{E2ETestSuite: s}
	suite.Run(t, testSuite)
}

func (s *FixTestSuite) TestFixRemovedMsgTypeQueryPanic() {
	t := s.T()
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	user := s.GetAndFundTestUser("default", 10_000_000_000, s.Chain)

	// v27 params update
	proposalID := s.SubmitBuilderParamsUpdate(s.Chain, user, e2esuite.DefaultAuthority)
	proposalIDInt, err := strconv.ParseUint(proposalID, 10, 64)
	require.NoError(t, err, "error parsing proposal ID as uint64")
	s.VoteOnProp(s.Chain, proposalIDInt, 0)
	t.Log("v27 params update proposal ID", proposalIDInt)

	// upgrade to v28
	height, err := s.Chain.Height(s.Ctx)
	require.NoError(t, err, "error fetching height before submit upgrade proposal")
	haltHeight := height + haltHeightDelta
	proposalID = s.SubmitSoftwareUpgradeProposal(s.Chain, user, v28UpgradeName, haltHeight, e2esuite.DefaultAuthority)
	proposalIDInt, err = strconv.ParseUint(proposalID, 10, 64)
	require.NoError(t, err, "failed to parse proposal ID")
	s.ValidatorVoting(s.Chain, proposalIDInt, height, haltHeight)
	s.UpgradeNodes(s.Chain, s.DockerClient, haltHeight, v28Chain.Repository, v28Chain.Version)
	t.Log("v28 upgrade successful!")

	// upgrade to v29
	height, err = s.Chain.Height(s.Ctx)
	require.NoError(t, err, "error fetching height before submit upgrade proposal")
	haltHeight = height + haltHeightDelta
	proposalID = s.SubmitSoftwareUpgradeProposal(s.Chain, user, v29UpgradeName, haltHeight, e2esuite.DefaultAuthority)
	proposalIDInt, err = strconv.ParseUint(proposalID, 10, 64)
	require.NoError(t, err, "failed to parse proposal ID")
	s.ValidatorVoting(s.Chain, proposalIDInt, height, haltHeight)
	s.UpgradeNodes(s.Chain, s.DockerClient, haltHeight, v29Chain.Repository, v29Chain.Version)
	t.Log("v29 upgrade successful!")

	// query gov module to check for panic
	proposals, err := s.Chain.GovQueryProposalsV1(s.Ctx, govv1types.ProposalStatus_PROPOSAL_STATUS_PASSED)
	require.NoError(t, err, "error querying gov module")
	t.Log("proposals", proposals)
}
