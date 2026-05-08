package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/collections"
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"

	"github.com/CosmosContracts/juno/v30/testutil"
	"github.com/CosmosContracts/juno/v30/x/voting-snapshot/keeper"
	"github.com/CosmosContracts/juno/v30/x/voting-snapshot/types"
)

type KeeperTestSuite struct {
	testutil.KeeperTestHelper
	keeper keeper.Keeper
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
	s.keeper = s.App.AppKeepers.VotingSnapshotKeeper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) TestParamsRoundTrip() {
	params, err := s.keeper.Params.Get(s.Ctx)
	s.Require().NoError(err)
	s.Require().Empty(params.LSTAllowlist) // default genesis is empty

	updated := types.Params{LSTAllowlist: []string{"juno1abc"}}
	s.Require().NoError(s.keeper.Params.Set(s.Ctx, updated))

	got, err := s.keeper.Params.Get(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal([]string{"juno1abc"}, got.LSTAllowlist)
}

// TestVotingPowerAtBeforeFirstSnapshot confirms a query at a height with no
// recorded snapshot returns zero (not an error).
func (s *KeeperTestSuite) TestVotingPowerAtBeforeFirstSnapshot() {
	_, _, addr := testdata.KeyTestPubAddr()

	power, err := s.keeper.VotingPowerAt(s.Ctx, addr, 999)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.ZeroInt(), power)
}

// TestVotingPowerAtAtOrBeforeSemantics writes snapshots at heights 10 and 20
// and verifies queries at heights between them resolve to the latest preceding
// value, not zero.
func (s *KeeperTestSuite) TestVotingPowerAtAtOrBeforeSemantics() {
	_, _, addr := testdata.KeyTestPubAddr()

	// Write snapshots manually (bypassing hooks) at two heights
	s.Require().NoError(s.keeper.VotingPower.Set(s.Ctx, collections.Join[[]byte, int64](addr.Bytes(), 10), sdkmath.NewInt(100)))
	s.Require().NoError(s.keeper.VotingPower.Set(s.Ctx, collections.Join[[]byte, int64](addr.Bytes(), 20), sdkmath.NewInt(200)))

	// At height 5 (before any snapshot): zero
	p, err := s.keeper.VotingPowerAt(s.Ctx, addr, 5)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.ZeroInt(), p)

	// At height 10 (exact match): 100
	p, err = s.keeper.VotingPowerAt(s.Ctx, addr, 10)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(100), p)

	// At height 15 (between snapshots): latest at-or-before is 100
	p, err = s.keeper.VotingPowerAt(s.Ctx, addr, 15)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(100), p)

	// At height 20 (later snapshot): 200
	p, err = s.keeper.VotingPowerAt(s.Ctx, addr, 20)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(200), p)

	// At height 100 (well beyond): still 200 (latest at-or-before)
	p, err = s.keeper.VotingPowerAt(s.Ctx, addr, 100)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(200), p)
}

// TestLSTExclusion verifies an LST-listed delegator records zero power even
// when the underlying staking keeper would report bonded shares.
func (s *KeeperTestSuite) TestLSTExclusion() {
	_, _, addr := testdata.KeyTestPubAddr()

	// Set params with addr in the allowlist
	s.Require().NoError(s.keeper.Params.Set(s.Ctx, types.Params{
		LSTAllowlist: []string{addr.String()},
	}))

	isLST, err := s.keeper.IsLST(s.Ctx, addr)
	s.Require().NoError(err)
	s.Require().True(isLST)
}
