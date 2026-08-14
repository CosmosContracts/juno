package keeper_test

import (
	"context"
	"encoding/binary"
	"fmt"
	"testing"

	coreheader "cosmossdk.io/core/header"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	snapshotkeeper "github.com/CosmosContracts/juno/v31/x/voting-snapshot/keeper"
	snapshottypes "github.com/CosmosContracts/juno/v31/x/voting-snapshot/types"
)

const benchmarkDelegators = 10_001

type benchmarkStakingKeeper struct {
	delegations []stakingtypes.Delegation
	byDelegator map[string]stakingtypes.Delegation
	validator   stakingtypes.Validator
	total       math.Int

	validatorDelegationIterations int64
	delegatorDelegationIterations int64
	validatorLookups              int64
}

func (m *benchmarkStakingKeeper) TotalBondedTokens(context.Context) (math.Int, error) {
	return m.total, nil
}

func (m *benchmarkStakingKeeper) IterateAllDelegations(_ context.Context, fn func(stakingtypes.Delegation) bool) error {
	for _, delegation := range m.delegations {
		if fn(delegation) {
			break
		}
	}
	return nil
}

func (m *benchmarkStakingKeeper) GetValidatorDelegations(context.Context, sdk.ValAddress) ([]stakingtypes.Delegation, error) {
	m.validatorDelegationIterations += int64(len(m.delegations))
	return m.delegations, nil
}

func (m *benchmarkStakingKeeper) IterateDelegatorDelegations(_ context.Context, delegator sdk.AccAddress, fn func(stakingtypes.Delegation) bool) error {
	delegation, ok := m.byDelegator[delegator.String()]
	if !ok {
		return nil
	}
	m.delegatorDelegationIterations++
	fn(delegation)
	return nil
}

func (m *benchmarkStakingKeeper) GetValidator(context.Context, sdk.ValAddress) (stakingtypes.Validator, error) {
	m.validatorLookups++
	return m.validator, nil
}

func deterministicAddress(index int) sdk.AccAddress {
	address := make([]byte, 20)
	binary.BigEndian.PutUint64(address[12:], uint64(index+1))
	return sdk.AccAddress(address)
}

func newBenchmarkVotingSnapshotKeeper(tb testing.TB, delegatorCount int) (sdk.Context, snapshotkeeper.Keeper, *benchmarkStakingKeeper, sdk.ValAddress) {
	tb.Helper()

	keys := storetypes.NewKVStoreKeys(snapshottypes.StoreKey)
	tkeys := storetypes.NewTransientStoreKeys(snapshottypes.TransientStoreKey)
	ctx := sdktestutil.DefaultContextWithKeys(keys, tkeys, nil).
		WithBlockHeight(31).
		WithHeaderInfo(coreheader.Info{Height: 31})

	valAddress := sdk.ValAddress(deterministicAddress(0x7fff))
	delegatorShares := math.LegacyNewDec(int64(2 * delegatorCount))
	validator := stakingtypes.Validator{
		OperatorAddress: valAddress.String(),
		Status:          stakingtypes.Bonded,
		Tokens:          math.NewInt(int64(2 * delegatorCount)),
		DelegatorShares: delegatorShares,
	}
	stakingKeeper := &benchmarkStakingKeeper{
		delegations: make([]stakingtypes.Delegation, 0, delegatorCount),
		byDelegator: make(map[string]stakingtypes.Delegation, delegatorCount),
		validator:   validator,
		total:       validator.Tokens,
	}
	for i := range delegatorCount {
		delegator := deterministicAddress(i)
		delegation := stakingtypes.Delegation{
			DelegatorAddress: delegator.String(),
			ValidatorAddress: valAddress.String(),
			Shares:           math.LegacyNewDec(2),
		}
		stakingKeeper.delegations = append(stakingKeeper.delegations, delegation)
		stakingKeeper.byDelegator[delegator.String()] = delegation
	}

	encoding := moduletestutil.MakeTestEncodingConfig()
	keeper := snapshotkeeper.NewKeeper(
		encoding.Codec,
		runtime.NewKVStoreService(keys[snapshottypes.StoreKey]),
		stakingKeeper,
		"authority",
		snapshottypes.NewTransientKVStoreService(tkeys[snapshottypes.TransientStoreKey]),
	)
	if err := keeper.Params.Set(ctx, snapshottypes.DefaultParams()); err != nil {
		tb.Fatalf("set params: %v", err)
	}
	return ctx, keeper, stakingKeeper, valAddress
}

func TestValidatorWalkOverAdvisoryLimitPreservesSameBlockPower(t *testing.T) {
	ctx, keeper, stakingKeeper, validator := newBenchmarkVotingSnapshotKeeper(t, benchmarkDelegators)

	// Model a 50%% slash after BeforeValidatorSlashed has marked every
	// delegator dirty. EndBlock must read the settled validator state.
	if err := keeper.Hooks().BeforeValidatorSlashed(ctx, validator, math.LegacyMustNewDecFromStr("0.5")); err != nil {
		t.Fatal(err)
	}
	stakingKeeper.validator.Tokens = math.NewInt(benchmarkDelegators)
	stakingKeeper.total = stakingKeeper.validator.Tokens
	if err := keeper.EndBlocker(ctx); err != nil {
		t.Fatal(err)
	}

	for _, index := range []int{0, benchmarkDelegators / 2, benchmarkDelegators - 1} {
		power, err := keeper.VotingPowerAt(ctx, deterministicAddress(index), ctx.BlockHeight())
		if err != nil {
			t.Fatal(err)
		}
		if !power.Equal(math.OneInt()) {
			t.Fatalf("delegator %d: got %s power, want 1", index, power)
		}
	}
	total, err := keeper.TotalVotingPowerAt(ctx, ctx.BlockHeight())
	if err != nil {
		t.Fatal(err)
	}
	if !total.Equal(math.NewInt(benchmarkDelegators)) {
		t.Fatalf("got total %s, want %d", total, benchmarkDelegators)
	}
	if stakingKeeper.validatorDelegationIterations != benchmarkDelegators {
		t.Fatalf("validator walk iterations = %d, want %d", stakingKeeper.validatorDelegationIterations, benchmarkDelegators)
	}
	if stakingKeeper.delegatorDelegationIterations != benchmarkDelegators {
		t.Fatalf("end-block delegator iterations = %d, want %d", stakingKeeper.delegatorDelegationIterations, benchmarkDelegators)
	}
	if stakingKeeper.validatorLookups != benchmarkDelegators {
		t.Fatalf("validator lookups = %d, want %d", stakingKeeper.validatorLookups, benchmarkDelegators)
	}
}

func BenchmarkValidatorDelegatorWalk(b *testing.B) {
	for _, count := range []int{1_000, 6_000, benchmarkDelegators} {
		b.Run(fmt.Sprintf("slash/%d", count), func(b *testing.B) {
			ctx, keeper, stakingKeeper, validator := newBenchmarkVotingSnapshotKeeper(b, count)
			b.ReportAllocs()
			b.ReportMetric(float64(count), "delegators/op")
			b.ResetTimer()
			for range b.N {
				if err := keeper.Hooks().BeforeValidatorSlashed(ctx, validator, math.LegacyZeroDec()); err != nil {
					b.Fatal(err)
				}
				if err := keeper.EndBlocker(ctx); err != nil {
					b.Fatal(err)
				}
			}
			b.StopTimer()
			b.ReportMetric(float64(stakingKeeper.validatorDelegationIterations)/float64(b.N), "validator-iterations/op")
			b.ReportMetric(float64(stakingKeeper.delegatorDelegationIterations)/float64(b.N), "endblock-iterations/op")
		})

		b.Run(fmt.Sprintf("status/%d", count), func(b *testing.B) {
			ctx, keeper, stakingKeeper, validator := newBenchmarkVotingSnapshotKeeper(b, count)
			b.ReportAllocs()
			b.ReportMetric(float64(count), "delegators/op")
			b.ResetTimer()
			for range b.N {
				if err := keeper.Hooks().AfterValidatorBeginUnbonding(ctx, nil, validator); err != nil {
					b.Fatal(err)
				}
				if err := keeper.EndBlocker(ctx); err != nil {
					b.Fatal(err)
				}
			}
			b.StopTimer()
			b.ReportMetric(float64(stakingKeeper.validatorDelegationIterations)/float64(b.N), "validator-iterations/op")
			b.ReportMetric(float64(stakingKeeper.delegatorDelegationIterations)/float64(b.N), "endblock-iterations/op")
		})
	}
}
