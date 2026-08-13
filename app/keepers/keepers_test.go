package keepers_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/CosmosContracts/juno/v31/testutil"
)

type KeepersTestSuite struct {
	testutil.KeeperTestHelper
}

func TestKeepersTestSuite(t *testing.T) {
	suite.Run(t, new(KeepersTestSuite))
}

// TestVotingSnapshotKeeperConstructedBeforeWasmPlugins is the C1 regression
// test: the wasm query plugin captures VotingSnapshotKeeper BY VALUE during
// NewAppKeepers, so the keeper must be fully constructed before
// RegisterCustomPlugins runs. NewAppKeepers panics if the keeper is still the
// zero value at that point (see the guard in keepers.go) — this test both
// exercises that path (app construction) and asserts the final keeper is
// usable.
func (s *KeepersTestSuite) TestVotingSnapshotKeeperConstructedBeforeWasmPlugins() {
	s.Setup()

	// a zero-value keeper has an empty authority and a nil schema; the
	// constructed keeper must have both populated
	s.Require().NotEmpty(s.App.AppKeepers.VotingSnapshotKeeper.Authority())

	// keeper reads must work (a zero-value keeper's collections panic)
	s.Require().NotPanics(func() {
		_, err := s.App.AppKeepers.VotingSnapshotKeeper.Params.Has(s.Ctx)
		s.Require().NoError(err)
	})
}
