package basic_test

import (
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/suite"

	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"
)

type BasicTestSuite struct {
	*e2esuite.E2ETestSuite
}

func TestBasicTestSuite(t *testing.T) {
	s := e2esuite.NewE2ETestSuite(
		[]*interchaintest.ChainSpec{e2esuite.DefaultSpec},
		e2esuite.DefaultTxCfg,
	)

	t.Parallel()
	t.Cleanup(func() {
		_ = s.Ic.Close()
	})

	testSuite := &BasicTestSuite{E2ETestSuite: s}
	suite.Run(t, testSuite)
}

// TestBasicStart is a basic test to assert that spinning up a chain with one validator works properly.
func (s *BasicTestSuite) TestBasicStart() {
	t := s.T()
	require := s.Require()
	if testing.Short() {
		t.Skip()
	}

	if err := testutil.WaitForBlocks(s.Ctx, 1, s.Chain); err != nil {
		require.NoError(err)
	}

	user := s.GetAndFundTestUser(t.Name(), 1_000_000_000, s.Chain)
	s.ConformanceCosmWasm(s.Chain, user)

	require.NotNil(s.Ic)
	require.NotNil(s.Ctx)
}
