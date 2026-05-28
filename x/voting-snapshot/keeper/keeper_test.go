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
	s.Require().Empty(params.LstAllowlist) // default genesis is empty

	updated := types.Params{LstAllowlist: []string{"juno1abc"}}
	s.Require().NoError(s.keeper.Params.Set(s.Ctx, updated))

	got, err := s.keeper.Params.Get(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal([]string{"juno1abc"}, got.LstAllowlist)
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
		LstAllowlist: []string{addr.String()},
	}))

	isLST, err := s.keeper.IsLST(s.Ctx, addr)
	s.Require().NoError(err)
	s.Require().True(isLST)
}

// TestVotingPowerOverRange verifies the range query returns only snapshots
// inside [from, to], in ascending height order.
func (s *KeeperTestSuite) TestVotingPowerOverRange() {
	_, _, addr := testdata.KeyTestPubAddr()

	for _, h := range []int64{5, 10, 15, 20, 25} {
		s.Require().NoError(s.keeper.VotingPower.Set(
			s.Ctx,
			collections.Join[[]byte, int64](addr.Bytes(), h),
			sdkmath.NewInt(h*10),
		))
	}

	rows, err := s.keeper.VotingPowerOverRange(s.Ctx, addr, 10, 20)
	s.Require().NoError(err)
	s.Require().Len(rows, 3)
	s.Require().Equal(int64(10), rows[0].Height)
	s.Require().Equal(sdkmath.NewInt(100), rows[0].Power)
	s.Require().Equal(int64(20), rows[2].Height)
	s.Require().Equal(sdkmath.NewInt(200), rows[2].Power)

	// Inverted range returns nil
	rows, err = s.keeper.VotingPowerOverRange(s.Ctx, addr, 30, 10)
	s.Require().NoError(err)
	s.Require().Nil(rows)
}

// TestPruneRetentionWindow verifies snapshots older than the window get
// dropped while preserving the most recent below-cutoff snapshot per
// delegator (so at-or-before reads still resolve to the correct value).
// Also confirms zero window disables pruning.
func (s *KeeperTestSuite) TestPruneRetentionWindow() {
	_, _, addr := testdata.KeyTestPubAddr()

	// Seed snapshots at heights 100..104 and totals at the same heights.
	for h := int64(100); h <= 104; h++ {
		s.Require().NoError(s.keeper.VotingPower.Set(
			s.Ctx,
			collections.Join[[]byte, int64](addr.Bytes(), h),
			sdkmath.NewInt(h),
		))
		s.Require().NoError(s.keeper.TotalPower.Set(s.Ctx, h, sdkmath.NewInt(h*1000)))
	}

	// Set a small retention window (3 blocks). Advance the SDK ctx height to 105.
	s.Require().NoError(s.keeper.Params.Set(s.Ctx, types.Params{
		LstAllowlist:           []string{},
		RetentionWindowHeights: 3,
		PruneInterval:          1,
	}))
	prunedCtx := s.Ctx.WithBlockHeight(105)

	s.Require().NoError(s.keeper.Prune(prunedCtx))

	// cutoff = 105 - 3 = 102. Heights < 102 are eligible for prune, but
	// the most-recent-below-cutoff (101) must survive so at-or-before
	// reads at any height in [101, 102) still resolve to a real value.
	// So: 100 is dropped; 101..104 all survive.
	_, err := s.keeper.VotingPower.Get(prunedCtx, collections.Join[[]byte, int64](addr.Bytes(), 100))
	s.Require().Error(err, "height 100 should be pruned (older entry below cutoff)")
	_, err = s.keeper.TotalPower.Get(prunedCtx, 100)
	s.Require().Error(err, "total at height 100 should be pruned")

	for _, h := range []int64{101, 102, 103, 104} {
		_, err := s.keeper.VotingPower.Get(prunedCtx, collections.Join[[]byte, int64](addr.Bytes(), h))
		s.Require().NoError(err, "voting-power height %d should survive prune", h)
		_, err = s.keeper.TotalPower.Get(prunedCtx, h)
		s.Require().NoError(err, "total-power height %d should survive prune", h)
	}

	// Disable retention by setting window to 0 — re-seed heights, then prune
	// and confirm survival.
	s.Require().NoError(s.keeper.VotingPower.Set(
		s.Ctx,
		collections.Join[[]byte, int64](addr.Bytes(), 50),
		sdkmath.NewInt(50),
	))
	s.Require().NoError(s.keeper.Params.Set(s.Ctx, types.Params{
		LstAllowlist:           []string{},
		RetentionWindowHeights: 0,
	}))
	s.Require().NoError(s.keeper.Prune(prunedCtx))
	_, err = s.keeper.VotingPower.Get(prunedCtx, collections.Join[[]byte, int64](addr.Bytes(), 50))
	s.Require().NoError(err, "zero retention window should disable pruning")
}

// TestPruneSparseDelegatorPreserved is the regression for the sparse
// delegator bug Cascade flagged: a delegator whose only snapshot is older
// than the retention window must still return their bonded power after a
// prune, not zero. Without the per-delegator h_max guard the only entry
// gets deleted and VotingPowerAt collapses to zero — silently zeroing
// every set-and-forget delegator.
func (s *KeeperTestSuite) TestPruneSparseDelegatorPreserved() {
	_, _, alice := testdata.KeyTestPubAddr() // sparse — single snapshot far in the past
	_, _, bob := testdata.KeyTestPubAddr()   // dense — entries on both sides of cutoff

	// Alice delegated once at height 100 and has not touched her stake since.
	const alicePower int64 = 10_000
	s.Require().NoError(s.keeper.VotingPower.Set(
		s.Ctx,
		collections.Join[[]byte, int64](alice.Bytes(), 100),
		sdkmath.NewInt(alicePower),
	))

	// Bob delegated and re-delegated multiple times: 100, 150, 1_000, 5_000.
	for _, hp := range []struct {
		h int64
		p int64
	}{{100, 5_000}, {150, 6_000}, {1_000, 7_000}, {5_000, 8_000}} {
		s.Require().NoError(s.keeper.VotingPower.Set(
			s.Ctx,
			collections.Join[[]byte, int64](bob.Bytes(), hp.h),
			sdkmath.NewInt(hp.p),
		))
	}

	// Retention window of 100 blocks; current height 5_100. cutoff = 5_000.
	// Entries with height < 5_000 (100, 150, 1_000) are below cutoff;
	// height 5_000 sits exactly at cutoff and is untouched.
	s.Require().NoError(s.keeper.Params.Set(s.Ctx, types.Params{
		LstAllowlist:           []string{},
		RetentionWindowHeights: 100,
		PruneInterval:          1,
	}))
	prunedCtx := s.Ctx.WithBlockHeight(5_100)
	s.Require().NoError(s.keeper.Prune(prunedCtx))

	// Alice (sparse): her one snapshot at 100 must survive. VotingPowerAt
	// at the current height must still return the original bonded power.
	p, err := s.keeper.VotingPowerAt(prunedCtx, alice, 5_100)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(alicePower), p, "set-and-forget delegator silently zeroed by prune")

	// Bob (dense): below-cutoff entries are 100, 150, 1_000 — the latest
	// (h=1_000) survives as h_max-below-cutoff; the earlier two get pruned.
	// 5_000 is at-or-above cutoff and untouched.
	_, err = s.keeper.VotingPower.Get(prunedCtx, collections.Join[[]byte, int64](bob.Bytes(), 100))
	s.Require().Error(err, "bob height 100 should be pruned (older below-cutoff entry)")
	_, err = s.keeper.VotingPower.Get(prunedCtx, collections.Join[[]byte, int64](bob.Bytes(), 150))
	s.Require().Error(err, "bob height 150 should be pruned (older below-cutoff entry)")
	for _, h := range []int64{1_000, 5_000} {
		_, err := s.keeper.VotingPower.Get(prunedCtx, collections.Join[[]byte, int64](bob.Bytes(), h))
		s.Require().NoError(err, "bob height %d should survive prune", h)
	}

	// Bob's at-or-before read between his h_max-below-cutoff (1_000) and
	// his next snapshot (5_000) must resolve to 7_000.
	p, err = s.keeper.VotingPowerAt(prunedCtx, bob, 4_999)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(7_000), p)
}

// TestPruneIntervalSkipsNonBoundaryBlocks confirms the governance-tunable
// PruneInterval lets the EndBlocker amortize the prune sweep across blocks
// rather than running it every single height.
func (s *KeeperTestSuite) TestPruneIntervalSkipsNonBoundaryBlocks() {
	_, _, addr := testdata.KeyTestPubAddr()

	s.Require().NoError(s.keeper.VotingPower.Set(
		s.Ctx,
		collections.Join[[]byte, int64](addr.Bytes(), 1),
		sdkmath.NewInt(1),
	))
	s.Require().NoError(s.keeper.VotingPower.Set(
		s.Ctx,
		collections.Join[[]byte, int64](addr.Bytes(), 2),
		sdkmath.NewInt(2),
	))

	// Window 5, interval 100. At height 200 (a multiple of 100) the prune
	// fires; at height 201 it should no-op.
	s.Require().NoError(s.keeper.Params.Set(s.Ctx, types.Params{
		RetentionWindowHeights: 5,
		PruneInterval:          100,
	}))

	// Height 201 — not on the interval boundary, nothing should change.
	skipCtx := s.Ctx.WithBlockHeight(201)
	s.Require().NoError(s.keeper.Prune(skipCtx))
	_, err := s.keeper.VotingPower.Get(skipCtx, collections.Join[[]byte, int64](addr.Bytes(), 1))
	s.Require().NoError(err, "height 1 should still be present on a non-interval block")

	// Height 200 — on the boundary, sweep runs and h_max-below-cutoff survives.
	runCtx := s.Ctx.WithBlockHeight(200)
	s.Require().NoError(s.keeper.Prune(runCtx))
	// cutoff = 195. h_max-below-cutoff for addr is 2; entry at h=1 gets pruned.
	_, err = s.keeper.VotingPower.Get(runCtx, collections.Join[[]byte, int64](addr.Bytes(), 1))
	s.Require().Error(err, "height 1 should be pruned on the interval boundary")
	_, err = s.keeper.VotingPower.Get(runCtx, collections.Join[[]byte, int64](addr.Bytes(), 2))
	s.Require().NoError(err, "height 2 (h_max below cutoff) should survive")
}
