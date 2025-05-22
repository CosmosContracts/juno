package basic

import (
	"testing"

	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"
	"github.com/stretchr/testify/suite"
)

type BasicTestSuite struct {
	e2esuite.E2ETestSuite
}

func TestBasicTestSuite(t *testing.T) {
	s := e2esuite.NewE2ETestSuite(
		e2esuite.Spec,
		e2esuite.TxCfg,
	)

	suite.Run(t, s)
}

// TestBasicStart is a basic test to assert that spinning up a chain with one validator works properly.
func (s *BasicTestSuite) TestBasicStart() {
	if testing.Short() {
		s.T().Skip()
	}

	s.T().Parallel()

	e2esuite.ConformanceCosmWasm(s.T(), s.Ctx, s.Chain, s.User1)

	s.Require().NotNil(s.Ic)
	s.Require().NotNil(s.Ctx)

	s.T().Cleanup(func() {
		_ = s.Ic.Close()
	})
}
