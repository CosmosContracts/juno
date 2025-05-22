package feemarket

import (
	"testing"

	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type FeemarketTestSuite struct {
	e2esuite.E2ETestSuite
}

func TestFeemarketTestSuite(t *testing.T) {
	s := e2esuite.NewE2ETestSuite(
		e2esuite.Spec,
		e2esuite.TxCfg,
	)

	suite.Run(t, s)
}

func (s *FeemarketTestSuite) TestQueryParams() {
	s.Run("query params", func() {
		// query params
		params := s.QueryParams()

		// expect validate to pass
		require.NoError(s.T(), params.ValidateBasic(), params)
	})
}

func (s *FeemarketTestSuite) TestQueryState() {
	s.Run("query state", func() {
		// query state
		state := s.QueryState()

		// expect validate to pass
		require.NoError(s.T(), state.ValidateBasic(), state)
	})
}

func (s *FeemarketTestSuite) TestQueryGasPrice() {
	s.Run("query gas price", func() {
		// query gas price
		gasPrice := s.QueryDefaultGasPrice()

		// expect validate to pass
		require.NoError(s.T(), gasPrice.Validate(), gasPrice)
	})
}
