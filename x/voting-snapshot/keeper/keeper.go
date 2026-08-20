package keeper

import (
	"context"

	"cosmossdk.io/collections"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v31/x/voting-snapshot/types"
)

// Keeper holds the event-driven voting-power index for staked JUNO.
//
// Per planning/05-staking-snapshot.md (option 2): staking hooks mark
// affected delegators dirty in a transient store, and the module
// EndBlocker (ordered after staking in app/modules.go) recomputes and
// writes their snapshots once all staking mutations for the block have
// settled. Reads return the latest snapshot at-or-before the requested
// height — caller-side semantics line up with proposal-vote tallying
// ("what was X's power at proposal-open height?").
//
// Writing from EndBlock instead of inside the hooks is load-bearing for
// correctness: several staking hooks (BeforeDelegationRemoved,
// BeforeValidatorSlashed) fire while the store still holds the
// pre-mutation state, so an in-hook recompute records stale power.
type Keeper struct {
	cdc           codec.BinaryCodec
	authority     string
	stakingKeeper types.StakingKeeper

	Schema collections.Schema
	Params collections.Item[types.Params]
	// VotingPower indexes a delegator's bonded stake per snapshot height.
	// Only delegations to validators in Bonded status count (matching the
	// TotalBondedTokens basis). LST-allowlisted delegators write zero and
	// their stake is symmetrically excluded from TotalPower.
	VotingPower collections.Map[collections.Pair[[]byte, int64], math.Int]
	// TotalPower indexes the chain's total voting power per snapshot
	// height: staking.TotalBondedTokens minus the bonded stake held by
	// LST-allowlisted addresses. This keeps the numerator/denominator
	// bases aligned so Σ VotingPower[d,h] <= TotalPower[h] (up to
	// shares-rounding dust). DAO designers can compute quorum as
	// Σ votes / TotalPower directly.
	TotalPower collections.Map[int64, math.Int]

	// TransientSchema and the collections below live in the module's
	// transient store (reset on every commit). They carry the set of
	// delegators whose power changed within the current block from the
	// staking hooks to the EndBlocker drain.
	TransientSchema collections.Schema
	DirtyDelegators collections.KeySet[[]byte]
	TotalDirty      collections.Item[bool]
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService corestore.KVStoreService,
	stakingKeeper types.StakingKeeper,
	authority string,
	transientService corestore.KVStoreService,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	tsb := collections.NewSchemaBuilder(transientService)

	k := Keeper{
		cdc:           cdc,
		authority:     authority,
		stakingKeeper: stakingKeeper,
		Params: collections.NewItem(
			sb,
			types.ParamsKey,
			"params",
			codec.CollValue[types.Params](cdc),
		),
		VotingPower: collections.NewMap(
			sb,
			types.VotingPowerKey,
			"voting_power",
			collections.PairKeyCodec(collections.BytesKey, collections.Int64Key),
			sdk.IntValue,
		),
		TotalPower: collections.NewMap(
			sb,
			types.TotalPowerKey,
			"total_power",
			collections.Int64Key,
			sdk.IntValue,
		),
		DirtyDelegators: collections.NewKeySet(
			tsb,
			types.TransientDirtyDelegatorsKey,
			"dirty_delegators",
			collections.BytesKey,
		),
		TotalDirty: collections.NewItem(
			tsb,
			types.TransientTotalDirtyKey,
			"total_dirty",
			collections.BoolValue,
		),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	tschema, err := tsb.Build()
	if err != nil {
		panic(err)
	}
	k.TransientSchema = tschema

	return k
}

func (k Keeper) Authority() string { return k.authority }

// Logger returns a module-tagged logger derived from the context.
func (Keeper) Logger(ctx context.Context) log.Logger {
	return sdk.UnwrapSDKContext(ctx).Logger().With("module", "x/"+types.ModuleName)
}

// IsLST reports whether the given delegator address is currently in the
// LST allowlist (delegators whose stake does NOT count toward voting power).
// Pre-genesis (params unset) returns false so staking hooks fired during
// InitGenesis don't error out before our InitGenesis sets params.
func (k Keeper) IsLST(ctx context.Context, addr sdk.AccAddress) (bool, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	bech := addr.String()
	for _, listed := range params.LstAllowlist {
		if listed == bech {
			return true, nil
		}
	}
	return false, nil
}
