package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/collections"
	coreheader "cosmossdk.io/core/header"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/CosmosContracts/juno/v31/x/voting-snapshot/keeper"
	"github.com/CosmosContracts/juno/v31/x/voting-snapshot/types"
)

const mintModuleName = "mint" // faucet module account for funding test delegators

// KeeperTestSuite wires the voting-snapshot keeper against REAL SDK
// auth/bank/staking keepers over an in-memory multistore (persistent +
// transient), with the module's staking hooks registered — so tests
// exercise the actual SDK v0.53 hook firing order (e.g.
// BeforeDelegationRemoved firing while the delegation is still in the
// store) rather than a mock's approximation of it.
type KeeperTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	accountKeeper authkeeper.AccountKeeper
	bankKeeper    bankkeeper.BaseKeeper
	stakingKeeper *stakingkeeper.Keeper
	keeper        keeper.Keeper

	bondDenom string
	authority string
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey, types.StoreKey,
	)
	tkeys := storetypes.NewTransientStoreKeys(types.TransientStoreKey)
	ctx := sdktestutil.DefaultContextWithKeys(keys, tkeys, nil)

	startTime := time.Unix(1_700_000_000, 0).UTC()
	ctx = ctx.WithBlockHeight(1).WithBlockTime(startTime).
		WithHeaderInfo(coreheader.Info{Height: 1, Time: startTime})

	encCfg := moduletestutil.MakeTestEncodingConfig(
		auth.AppModuleBasic{}, bank.AppModuleBasic{}, staking.AppModuleBasic{},
	)

	maccPerms := map[string][]string{
		authtypes.FeeCollectorName:     nil,
		mintModuleName:                 {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	}
	s.authority = authtypes.NewModuleAddress(govtypes.ModuleName).String()

	s.accountKeeper = authkeeper.NewAccountKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		address.NewBech32Codec(sdk.Bech32MainPrefix),
		sdk.Bech32MainPrefix,
		s.authority,
	)
	s.bankKeeper = bankkeeper.NewBaseKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		s.accountKeeper,
		map[string]bool{},
		s.authority,
		log.NewNopLogger(),
	)
	s.stakingKeeper = stakingkeeper.NewKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
		s.accountKeeper,
		s.bankKeeper,
		s.authority,
		address.NewBech32Codec(sdk.Bech32PrefixValAddr),
		address.NewBech32Codec(sdk.Bech32PrefixConsAddr),
	)
	s.Require().NoError(s.stakingKeeper.SetParams(ctx, stakingtypes.DefaultParams()))

	s.keeper = keeper.NewKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(keys[types.StoreKey]),
		s.stakingKeeper,
		s.authority,
		types.NewTransientKVStoreService(tkeys[types.TransientStoreKey]),
	)
	s.stakingKeeper.SetHooks(s.keeper.Hooks())

	s.Require().NoError(s.keeper.Params.Set(ctx, types.DefaultParams()))

	bondDenom, err := s.stakingKeeper.BondDenom(ctx)
	s.Require().NoError(err)
	s.bondDenom = bondDenom
	s.ctx = ctx
}

// ---- end-to-end staking-event tests --------------------------------------

const (
	valStake   int64 = 1_000_000 // 1 consensus power at default PowerReduction
	aliceStake int64 = 500_000
)

// TestDelegateRecordsPower: a fresh delegation is snapshotted at the end
// of the block it happened in, and history before it reads zero.
func (s *KeeperTestSuite) TestDelegateRecordsPower() {
	val, _ := s.createValidator()
	s.endBlock() // h1: validator bonds

	alice := s.newDelegator(aliceStake, val)
	s.endBlock() // h2

	s.Require().Equal(sdkmath.ZeroInt(), s.powerAt(alice, 1))
	s.Require().Equal(sdkmath.NewInt(aliceStake), s.powerAt(alice, 2))
	s.Require().Equal(sdkmath.NewInt(valStake+aliceStake), s.totalAt(2))
	// self-delegation of the validator operator is counted too
	s.Require().Equal(sdkmath.NewInt(valStake), s.powerAt(sdk.AccAddress(val), 2))
}

// TestPartialUndelegate: power drops by exactly the undelegated amount at
// the undelegation height.
func (s *KeeperTestSuite) TestPartialUndelegate() {
	val, _ := s.createValidator()
	s.endBlock() // h1

	alice := s.newDelegator(aliceStake, val)
	s.endBlock() // h2

	s.undelegate(alice, val, 200_000)
	s.endBlock() // h3

	s.Require().Equal(sdkmath.NewInt(aliceStake), s.powerAt(alice, 2))
	s.Require().Equal(sdkmath.NewInt(aliceStake-200_000), s.powerAt(alice, 3))
	s.Require().Equal(sdkmath.NewInt(valStake+aliceStake-200_000), s.totalAt(3))
}

// TestFullUndelegateZeroesPower is the C3 regression: SDK v0.53's
// BeforeDelegationRemoved hook fires while the store still holds the old
// delegation, so an in-hook recompute records the pre-removal amount and
// the delegator keeps phantom voting power forever. The EndBlocker drain
// must record zero at the undelegation height.
func (s *KeeperTestSuite) TestFullUndelegateZeroesPower() {
	val, _ := s.createValidator()
	s.endBlock() // h1

	alice := s.newDelegator(aliceStake, val)
	s.endBlock() // h2

	s.undelegate(alice, val, aliceStake) // full exit
	s.endBlock()                         // h3

	// history is preserved...
	s.Require().Equal(sdkmath.NewInt(aliceStake), s.powerAt(alice, 2))
	// ...but power at (and after) the undelegation height is ZERO.
	s.Require().Equal(sdkmath.ZeroInt(), s.powerAt(alice, 3), "full undelegation must zero voting power (C3 phantom-power regression)")
	s.Require().Equal(sdkmath.ZeroInt(), s.powerAt(alice, 100))
	s.Require().Equal(sdkmath.NewInt(valStake), s.totalAt(3))
}

// TestRedelegate: moving stake between two bonded validators keeps the
// delegator's power constant — no phantom dip or double count.
func (s *KeeperTestSuite) TestRedelegate() {
	val1, _ := s.createValidator()
	val2, _ := s.createValidator()
	s.endBlock() // h1

	alice := s.newDelegator(aliceStake, val1)
	s.endBlock() // h2

	shares, err := s.stakingKeeper.ValidateUnbondAmount(s.ctx, alice, val1, sdkmath.NewInt(aliceStake))
	s.Require().NoError(err)
	_, err = s.stakingKeeper.BeginRedelegation(s.ctx, alice, val1, val2, shares)
	s.Require().NoError(err)
	s.endBlock() // h3

	s.Require().Equal(sdkmath.NewInt(aliceStake), s.powerAt(alice, 2))
	s.Require().Equal(sdkmath.NewInt(aliceStake), s.powerAt(alice, 3))
	s.Require().Equal(sdkmath.NewInt(2*valStake+aliceStake), s.totalAt(3))
}

// TestSlashRecordsPostSlashPower is the H2 regression: SDK v0.53's
// BeforeValidatorSlashed fires before RemoveValidatorTokens, so an
// in-hook recompute records pre-slash values. The EndBlocker drain must
// record the post-slash power.
func (s *KeeperTestSuite) TestSlashRecordsPostSlashPower() {
	val, cons := s.createValidator()
	s.endBlock() // h1

	alice := s.newDelegator(valStake, val) // equal stake → clean halving math
	s.endBlock()                           // h2: validator holds 2_000_000 (power 2)

	validator, err := s.stakingKeeper.GetValidator(s.ctx, val)
	s.Require().NoError(err)
	power := validator.GetConsensusPower(s.stakingKeeper.PowerReduction(s.ctx))

	_, err = s.stakingKeeper.Slash(s.ctx, cons, s.ctx.BlockHeight(), power, sdkmath.LegacyMustNewDecFromStr("0.5"))
	s.Require().NoError(err)
	s.endBlock() // h3

	s.Require().Equal(sdkmath.NewInt(valStake), s.powerAt(alice, 2))
	s.Require().Equal(sdkmath.NewInt(valStake/2), s.powerAt(alice, 3), "slash must record post-slash power (H2 regression)")
	s.Require().Equal(sdkmath.NewInt(valStake/2), s.powerAt(sdk.AccAddress(val), 3))
	s.Require().Equal(sdkmath.NewInt(valStake), s.totalAt(3))
}

// TestJailedValidatorExcludedFromPower is the F4 regression: when a
// validator leaves the bonded set, TotalBondedTokens drops immediately
// but no delegation hooks fire — without the validator bond-status hooks
// its delegators' per-address power would stay stale (numerator >
// denominator basis). The AfterValidatorBeginUnbonding hook must dirty
// its delegators so their power is re-recorded on the same block.
func (s *KeeperTestSuite) TestJailedValidatorExcludedFromPower() {
	val1, _ := s.createValidator()
	val2, cons2 := s.createValidator()
	s.endBlock() // h1: both bond

	alice := s.newDelegator(aliceStake, val2)
	s.endBlock() // h2

	s.Require().NoError(s.stakingKeeper.Jail(s.ctx, cons2))
	s.endBlock() // h3: staking EndBlocker moves val2 to Unbonding

	h := int64(3)
	s.Require().Equal(sdkmath.NewInt(aliceStake), s.powerAt(alice, 2))
	s.Require().Equal(sdkmath.ZeroInt(), s.powerAt(alice, h), "delegation to a non-bonded validator must not count (F4 regression)")
	s.Require().Equal(sdkmath.ZeroInt(), s.powerAt(sdk.AccAddress(val2), h))
	s.Require().Equal(sdkmath.NewInt(valStake), s.powerAt(sdk.AccAddress(val1), h))
	s.Require().Equal(sdkmath.NewInt(valStake), s.totalAt(h))
}

// TestLSTExcludedFromPowerAndTotal (F5): allowlisted LST addresses record
// zero per-address power AND their stake is excluded from TotalPower —
// numerator and denominator stay on one basis, Σ power <= total.
func (s *KeeperTestSuite) TestLSTExcludedFromPowerAndTotal() {
	val, _ := s.createValidator()
	s.endBlock() // h1

	_, _, lst := testdata.KeyTestPubAddr()
	params := types.DefaultParams()
	params.LstAllowlist = []string{lst.String()}
	s.Require().NoError(s.keeper.Params.Set(s.ctx, params))

	s.fundAccount(lst, sdkmath.NewInt(aliceStake))
	s.delegate(lst, val, aliceStake)
	s.endBlock() // h2

	s.Require().Equal(sdkmath.ZeroInt(), s.powerAt(lst, 2), "LST power must be zero")
	s.Require().Equal(sdkmath.NewInt(valStake), s.totalAt(2), "LST stake must be excluded from TotalPower (F5 symmetry)")
}

// TestUpdateParamsValidatesAndRerecords: MsgUpdateParams rejects invalid
// params, and an allowlist change re-records the affected addresses and
// the total at the same height (no stale numerator/denominator split).
func (s *KeeperTestSuite) TestUpdateParamsValidatesAndRerecords() {
	val, _ := s.createValidator()
	s.endBlock() // h1

	alice := s.newDelegator(aliceStake, val)
	s.endBlock() // h2

	msgServer := keeper.NewMsgServer(s.keeper)

	// invalid params rejected
	_, err := msgServer.UpdateParams(s.ctx, &types.MsgUpdateParams{
		Authority: s.authority,
		Params:    types.Params{LstAllowlist: []string{"not-bech32"}},
	})
	s.Require().Error(err)

	// wrong authority rejected
	_, err = msgServer.UpdateParams(s.ctx, &types.MsgUpdateParams{
		Authority: alice.String(),
		Params:    types.DefaultParams(),
	})
	s.Require().Error(err)

	// listing alice as an LST zeroes her power and shrinks the total at
	// this height, without waiting for her next staking event
	newParams := types.DefaultParams()
	newParams.LstAllowlist = []string{alice.String()}
	_, err = msgServer.UpdateParams(s.ctx, &types.MsgUpdateParams{Authority: s.authority, Params: newParams})
	s.Require().NoError(err)
	s.endBlock() // h3

	s.Require().Equal(sdkmath.NewInt(aliceStake), s.powerAt(alice, 2))
	s.Require().Equal(sdkmath.ZeroInt(), s.powerAt(alice, 3))
	s.Require().Equal(sdkmath.NewInt(valStake), s.totalAt(3))

	// de-listing restores her power symmetrically
	_, err = msgServer.UpdateParams(s.ctx, &types.MsgUpdateParams{Authority: s.authority, Params: types.DefaultParams()})
	s.Require().NoError(err)
	s.endBlock() // h4

	s.Require().Equal(sdkmath.NewInt(aliceStake), s.powerAt(alice, 4))
	s.Require().Equal(sdkmath.NewInt(valStake+aliceStake), s.totalAt(4))
}

// TestEndBlockerDrainsDirtySet: the dirty set is cleared by the drain, so
// a block with no staking events writes no snapshots.
func (s *KeeperTestSuite) TestEndBlockerDrainsDirtySet() {
	val, _ := s.createValidator()
	s.endBlock() // h1

	count := 0
	s.Require().NoError(s.keeper.DirtyDelegators.Walk(s.ctx, nil, func(_ []byte) (bool, error) {
		count++
		return false, nil
	}))
	s.Require().Zero(count, "dirty set must be drained by EndBlocker")

	totalDirty, err := s.keeper.TotalDirty.Has(s.ctx)
	s.Require().NoError(err)
	s.Require().False(totalDirty)

	// a no-event block writes nothing new
	s.endBlock() // h2
	_, err = s.keeper.TotalPower.Get(s.ctx, 2)
	s.Require().Error(err, "no staking events => no snapshot at h2")
	// at-or-before read still resolves
	s.Require().Equal(sdkmath.NewInt(valStake), s.totalAt(2))
	_ = val
}

// ---- unit tests: params validation ---------------------------------------

func (s *KeeperTestSuite) TestParamsValidate() {
	_, _, addr := testdata.KeyTestPubAddr()

	valid := types.Params{
		LstAllowlist:           []string{addr.String()},
		RetentionWindowHeights: 12_614_400,
		PruneInterval:          100,
	}
	s.Require().NoError(valid.Validate())
	s.Require().NoError(types.DefaultParams().Validate())

	badBech := valid
	badBech.LstAllowlist = []string{"juno1notvalid"}
	s.Require().Error(badBech.Validate(), "invalid bech32 must be rejected")

	dup := valid
	dup.LstAllowlist = []string{addr.String(), addr.String()}
	s.Require().Error(dup.Validate(), "duplicate allowlist entries must be rejected")

	overflowRetention := valid
	overflowRetention.RetentionWindowHeights = ^uint64(0) // > MaxInt64
	s.Require().Error(overflowRetention.Validate())

	overflowInterval := valid
	overflowInterval.PruneInterval = ^uint64(0)
	s.Require().Error(overflowInterval.Validate())
}

// ---- unit tests: snapshot read semantics (store-level) --------------------

func (s *KeeperTestSuite) TestVotingPowerAtBeforeFirstSnapshot() {
	_, _, addr := testdata.KeyTestPubAddr()

	power, err := s.keeper.VotingPowerAt(s.ctx, addr, 999)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.ZeroInt(), power)
}

func (s *KeeperTestSuite) TestVotingPowerAtAtOrBeforeSemantics() {
	_, _, addr := testdata.KeyTestPubAddr()

	s.Require().NoError(s.keeper.VotingPower.Set(s.ctx, collections.Join[[]byte, int64](addr.Bytes(), 10), sdkmath.NewInt(100)))
	s.Require().NoError(s.keeper.VotingPower.Set(s.ctx, collections.Join[[]byte, int64](addr.Bytes(), 20), sdkmath.NewInt(200)))

	for _, tc := range []struct {
		at   int64
		want int64
	}{{5, 0}, {10, 100}, {15, 100}, {20, 200}, {100, 200}} {
		p, err := s.keeper.VotingPowerAt(s.ctx, addr, tc.at)
		s.Require().NoError(err)
		s.Require().Equal(sdkmath.NewInt(tc.want), p, "at height %d", tc.at)
	}
}

func (s *KeeperTestSuite) TestIsLST() {
	_, _, addr := testdata.KeyTestPubAddr()

	s.Require().NoError(s.keeper.Params.Set(s.ctx, types.Params{
		LstAllowlist: []string{addr.String()},
	}))

	isLST, err := s.keeper.IsLST(s.ctx, addr)
	s.Require().NoError(err)
	s.Require().True(isLST)
}

// ---- unit tests: range queries and caps -----------------------------------

func (s *KeeperTestSuite) TestVotingPowerOverRange() {
	_, _, addr := testdata.KeyTestPubAddr()

	for _, h := range []int64{5, 10, 15, 20, 25} {
		s.Require().NoError(s.keeper.VotingPower.Set(
			s.ctx,
			collections.Join[[]byte, int64](addr.Bytes(), h),
			sdkmath.NewInt(h*10),
		))
	}

	rows, err := s.keeper.VotingPowerOverRange(s.ctx, addr, 10, 20)
	s.Require().NoError(err)
	s.Require().Len(rows, 3)
	s.Require().Equal(int64(10), rows[0].Height)
	s.Require().Equal(sdkmath.NewInt(100), rows[0].Power)
	s.Require().Equal(int64(20), rows[2].Height)
	s.Require().Equal(sdkmath.NewInt(200), rows[2].Power)

	// inverted range returns nil
	rows, err = s.keeper.VotingPowerOverRange(s.ctx, addr, 30, 10)
	s.Require().NoError(err)
	s.Require().Nil(rows)
}

func (s *KeeperTestSuite) TestVotingPowerOverRangeCapped() {
	_, _, addr := testdata.KeyTestPubAddr()

	// normal case matches the uncapped variant
	for _, h := range []int64{5, 10, 15} {
		s.Require().NoError(s.keeper.VotingPower.Set(
			s.ctx,
			collections.Join[[]byte, int64](addr.Bytes(), h),
			sdkmath.NewInt(h),
		))
	}
	rows, err := s.keeper.VotingPowerOverRangeCapped(s.ctx, addr, 0, 100)
	s.Require().NoError(err)
	s.Require().Len(rows, 3)

	// over-wide window rejected
	_, err = s.keeper.VotingPowerOverRangeCapped(s.ctx, addr, 0, keeper.MaxVotingPowerRangeWidth+1)
	s.Require().ErrorIs(err, keeper.ErrRangeTooWide)

	// inverted range still returns nil, not an error
	rows, err = s.keeper.VotingPowerOverRangeCapped(s.ctx, addr, 100, 10)
	s.Require().NoError(err)
	s.Require().Nil(rows)

	// too many rows rejected (never truncated)
	_, _, dense := testdata.KeyTestPubAddr()
	for h := int64(1); h <= int64(keeper.MaxVotingPowerRangeRows)+1; h++ {
		s.Require().NoError(s.keeper.VotingPower.Set(
			s.ctx,
			collections.Join[[]byte, int64](dense.Bytes(), h),
			sdkmath.NewInt(h),
		))
	}
	_, err = s.keeper.VotingPowerOverRangeCapped(s.ctx, dense, 1, int64(keeper.MaxVotingPowerRangeRows)+1)
	s.Require().ErrorIs(err, keeper.ErrRangeTooManyRows)

	// narrowing the range succeeds
	rows, err = s.keeper.VotingPowerOverRangeCapped(s.ctx, dense, 1, int64(keeper.MaxVotingPowerRangeRows))
	s.Require().NoError(err)
	s.Require().Len(rows, keeper.MaxVotingPowerRangeRows)
}

// ---- unit tests: pruning ---------------------------------------------------

func (s *KeeperTestSuite) TestPruneRetentionWindow() {
	_, _, addr := testdata.KeyTestPubAddr()

	for h := int64(100); h <= 104; h++ {
		s.Require().NoError(s.keeper.VotingPower.Set(
			s.ctx,
			collections.Join[[]byte, int64](addr.Bytes(), h),
			sdkmath.NewInt(h),
		))
		s.Require().NoError(s.keeper.TotalPower.Set(s.ctx, h, sdkmath.NewInt(h*1000)))
	}

	s.Require().NoError(s.keeper.Params.Set(s.ctx, types.Params{
		LstAllowlist:           []string{},
		RetentionWindowHeights: 3,
		PruneInterval:          1,
	}))
	prunedCtx := s.ctx.WithBlockHeight(105)

	s.Require().NoError(s.keeper.Prune(prunedCtx))

	// cutoff = 102. Heights < 102 are eligible, but the most-recent
	// below-cutoff (101) must survive for at-or-before reads.
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

	// zero window disables pruning
	s.Require().NoError(s.keeper.VotingPower.Set(
		s.ctx,
		collections.Join[[]byte, int64](addr.Bytes(), 50),
		sdkmath.NewInt(50),
	))
	s.Require().NoError(s.keeper.Params.Set(s.ctx, types.Params{
		LstAllowlist:           []string{},
		RetentionWindowHeights: 0,
	}))
	s.Require().NoError(s.keeper.Prune(prunedCtx))
	_, err = s.keeper.VotingPower.Get(prunedCtx, collections.Join[[]byte, int64](addr.Bytes(), 50))
	s.Require().NoError(err, "zero retention window should disable pruning")
}

// TestPruneSparseDelegatorPreserved: a delegator whose only snapshot is
// older than the retention window must still resolve their power after a
// prune (h_max-below-cutoff guard).
func (s *KeeperTestSuite) TestPruneSparseDelegatorPreserved() {
	_, _, alice := testdata.KeyTestPubAddr() // sparse — single snapshot far in the past
	_, _, bob := testdata.KeyTestPubAddr()   // dense — entries on both sides of cutoff

	const alicePower int64 = 10_000
	s.Require().NoError(s.keeper.VotingPower.Set(
		s.ctx,
		collections.Join[[]byte, int64](alice.Bytes(), 100),
		sdkmath.NewInt(alicePower),
	))

	for _, hp := range []struct {
		h int64
		p int64
	}{{100, 5_000}, {150, 6_000}, {1_000, 7_000}, {5_000, 8_000}} {
		s.Require().NoError(s.keeper.VotingPower.Set(
			s.ctx,
			collections.Join[[]byte, int64](bob.Bytes(), hp.h),
			sdkmath.NewInt(hp.p),
		))
	}

	s.Require().NoError(s.keeper.Params.Set(s.ctx, types.Params{
		LstAllowlist:           []string{},
		RetentionWindowHeights: 100,
		PruneInterval:          1,
	}))
	prunedCtx := s.ctx.WithBlockHeight(5_100)
	s.Require().NoError(s.keeper.Prune(prunedCtx))

	p, err := s.keeper.VotingPowerAt(prunedCtx, alice, 5_100)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(alicePower), p, "set-and-forget delegator silently zeroed by prune")

	_, err = s.keeper.VotingPower.Get(prunedCtx, collections.Join[[]byte, int64](bob.Bytes(), 100))
	s.Require().Error(err, "bob height 100 should be pruned (older below-cutoff entry)")
	_, err = s.keeper.VotingPower.Get(prunedCtx, collections.Join[[]byte, int64](bob.Bytes(), 150))
	s.Require().Error(err, "bob height 150 should be pruned (older below-cutoff entry)")
	for _, h := range []int64{1_000, 5_000} {
		_, err := s.keeper.VotingPower.Get(prunedCtx, collections.Join[[]byte, int64](bob.Bytes(), h))
		s.Require().NoError(err, "bob height %d should survive prune", h)
	}

	p, err = s.keeper.VotingPowerAt(prunedCtx, bob, 4_999)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(7_000), p)
}

func (s *KeeperTestSuite) TestPruneIntervalSkipsNonBoundaryBlocks() {
	_, _, addr := testdata.KeyTestPubAddr()

	s.Require().NoError(s.keeper.VotingPower.Set(
		s.ctx,
		collections.Join[[]byte, int64](addr.Bytes(), 1),
		sdkmath.NewInt(1),
	))
	s.Require().NoError(s.keeper.VotingPower.Set(
		s.ctx,
		collections.Join[[]byte, int64](addr.Bytes(), 2),
		sdkmath.NewInt(2),
	))

	s.Require().NoError(s.keeper.Params.Set(s.ctx, types.Params{
		RetentionWindowHeights: 5,
		PruneInterval:          100,
	}))

	skipCtx := s.ctx.WithBlockHeight(201)
	s.Require().NoError(s.keeper.Prune(skipCtx))
	_, err := s.keeper.VotingPower.Get(skipCtx, collections.Join[[]byte, int64](addr.Bytes(), 1))
	s.Require().NoError(err, "height 1 should still be present on a non-interval block")

	runCtx := s.ctx.WithBlockHeight(200)
	s.Require().NoError(s.keeper.Prune(runCtx))
	_, err = s.keeper.VotingPower.Get(runCtx, collections.Join[[]byte, int64](addr.Bytes(), 1))
	s.Require().Error(err, "height 1 should be pruned on the interval boundary")
	_, err = s.keeper.VotingPower.Get(runCtx, collections.Join[[]byte, int64](addr.Bytes(), 2))
	s.Require().NoError(err, "height 2 (h_max below cutoff) should survive")
}

// TestPruneDeletionCap: a sweep deletes at most MaxPruneDeletionsPerRun
// keys and defers the remainder to the next interval — no unbounded
// delete burst, and nothing a read needs is lost.
func (s *KeeperTestSuite) TestPruneDeletionCap() {
	_, _, addr := testdata.KeyTestPubAddr()

	// One delegator, heights 1..cap+2 — all below cutoff. The pruner
	// keeps the most recent below-cutoff entry (cap+2), so cap+1 keys
	// are stale in total: one full run (cap deletions) plus a remainder.
	top := int64(keeper.MaxPruneDeletionsPerRun) + 2
	for h := int64(1); h <= top; h++ {
		s.Require().NoError(s.keeper.VotingPower.Set(
			s.ctx,
			collections.Join[[]byte, int64](addr.Bytes(), h),
			sdkmath.NewInt(1),
		))
	}

	s.Require().NoError(s.keeper.Params.Set(s.ctx, types.Params{
		RetentionWindowHeights: 100,
		PruneInterval:          1,
	}))

	count := func(ctx sdk.Context) int {
		n := 0
		rng := collections.NewPrefixedPairRange[[]byte, int64](addr.Bytes())
		s.Require().NoError(s.keeper.VotingPower.Walk(ctx, rng, func(_ collections.Pair[[]byte, int64], _ sdkmath.Int) (bool, error) {
			n++
			return false, nil
		}))
		return n
	}

	// run 1: capped at MaxPruneDeletionsPerRun deletions
	run1 := s.ctx.WithBlockHeight(top + 1_000)
	s.Require().NoError(s.keeper.Prune(run1))
	s.Require().Equal(int(top)-keeper.MaxPruneDeletionsPerRun, count(run1), "first run must delete exactly the cap")

	// run 2 (next interval): drains the remainder, keeping h_max below cutoff
	run2 := s.ctx.WithBlockHeight(top + 1_001)
	s.Require().NoError(s.keeper.Prune(run2))
	s.Require().Equal(1, count(run2), "second run must leave only the most-recent below-cutoff snapshot")

	p, err := s.keeper.VotingPowerAt(run2, addr, top+1_001)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(1), p, "at-or-before read must survive capped pruning")
}

// ---- block / staking helpers -------------------------------------------

// endBlock mirrors app/modules.go orderEndBlockers: staking's EndBlocker
// (validator-set updates, bond-status transitions) runs first, then the
// voting-snapshot EndBlocker drains the dirty set. Afterwards the context
// advances to the next height.
func (s *KeeperTestSuite) endBlock() {
	_, err := s.stakingKeeper.EndBlocker(s.ctx)
	s.Require().NoError(err)
	s.Require().NoError(s.keeper.EndBlocker(s.ctx))

	next := s.ctx.BlockHeight() + 1
	nextTime := s.ctx.BlockTime().Add(5 * time.Second)
	s.ctx = s.ctx.WithBlockHeight(next).WithBlockTime(nextTime).
		WithHeaderInfo(coreheader.Info{Height: next, Time: nextTime})
}

func (s *KeeperTestSuite) fundAccount(addr sdk.AccAddress, amt sdkmath.Int) {
	coins := sdk.NewCoins(sdk.NewCoin(s.bondDenom, amt))
	s.Require().NoError(s.bankKeeper.MintCoins(s.ctx, mintModuleName, coins))
	s.Require().NoError(s.bankKeeper.SendCoinsFromModuleToAccount(s.ctx, mintModuleName, addr, coins))
}

// createValidator registers a validator with a self-delegation of
// valStake and returns its operator + consensus addresses. The validator
// becomes Bonded at the next endBlock().
func (s *KeeperTestSuite) createValidator() (sdk.ValAddress, sdk.ConsAddress) {
	pk := ed25519.GenPrivKey().PubKey()
	valAddr := sdk.ValAddress(pk.Address())

	val, err := stakingtypes.NewValidator(valAddr.String(), pk, stakingtypes.Description{Moniker: "test-val"})
	s.Require().NoError(err)
	s.Require().NoError(s.stakingKeeper.SetValidator(s.ctx, val))
	s.Require().NoError(s.stakingKeeper.SetValidatorByConsAddr(s.ctx, val))
	s.Require().NoError(s.stakingKeeper.SetNewValidatorByPowerIndex(s.ctx, val))

	owner := sdk.AccAddress(valAddr)
	s.fundAccount(owner, sdkmath.NewInt(valStake))
	s.delegate(owner, valAddr, valStake)

	return valAddr, sdk.ConsAddress(pk.Address())
}

func (s *KeeperTestSuite) delegate(del sdk.AccAddress, val sdk.ValAddress, amt int64) {
	validator, err := s.stakingKeeper.GetValidator(s.ctx, val)
	s.Require().NoError(err)
	_, err = s.stakingKeeper.Delegate(s.ctx, del, sdkmath.NewInt(amt), stakingtypes.Unbonded, validator, true)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) undelegate(del sdk.AccAddress, val sdk.ValAddress, amt int64) {
	shares, err := s.stakingKeeper.ValidateUnbondAmount(s.ctx, del, val, sdkmath.NewInt(amt))
	s.Require().NoError(err)
	_, _, err = s.stakingKeeper.Undelegate(s.ctx, del, val, shares)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) newDelegator(stake int64, val sdk.ValAddress) sdk.AccAddress {
	_, _, addr := testdata.KeyTestPubAddr()
	s.fundAccount(addr, sdkmath.NewInt(stake))
	s.delegate(addr, val, stake)
	return addr
}

func (s *KeeperTestSuite) powerAt(del sdk.AccAddress, height int64) sdkmath.Int {
	p, err := s.keeper.VotingPowerAt(s.ctx, del, height)
	s.Require().NoError(err)
	return p
}

func (s *KeeperTestSuite) totalAt(height int64) sdkmath.Int {
	p, err := s.keeper.TotalVotingPowerAt(s.ctx, height)
	s.Require().NoError(err)
	return p
}
