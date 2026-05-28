package keeper

import (
	"context"

	"cosmossdk.io/collections"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v30/x/voting-snapshot/types"
)

// Keeper holds the event-driven voting-power index for staked JUNO.
//
// Per planning/05-staking-snapshot.md (option 2): writes happen on
// staking events. Reads return the latest snapshot at-or-before the
// requested height — caller-side semantics line up with proposal-vote
// tallying ("what was X's power at proposal-open height?").
type Keeper struct {
	cdc           codec.BinaryCodec
	authority     string
	stakingKeeper types.StakingKeeper

	Schema collections.Schema
	Params collections.Item[types.Params]
	// VotingPower indexes a delegator's bonded stake per snapshot height.
	// LST-allowlisted delegators write zero (their stake is excluded from
	// per-address voting power).
	VotingPower collections.Map[collections.Pair[[]byte, int64], math.Int]
	// TotalPower indexes the chain's total bonded stake per snapshot height.
	// Sourced from staking.TotalBondedTokens at write time, which still
	// includes LST bonded stake — so Σ VotingPower[d,h] < TotalPower[h] by
	// the LST share. Denominator subtraction is a planned v30.x refinement;
	// see planning/05-staking-snapshot.md "LST asymmetry" for the design
	// rationale. DAO designers computing quorum need to account for this.
	TotalPower collections.Map[int64, math.Int]
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService corestore.KVStoreService,
	stakingKeeper types.StakingKeeper,
	authority string,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)

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
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

func (k Keeper) Authority() string { return k.authority }

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
