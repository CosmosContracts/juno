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

	Schema      collections.Schema
	Params      collections.Item[types.Params]
	VotingPower collections.Map[collections.Pair[[]byte, int64], math.Int]
	TotalPower  collections.Map[int64, math.Int]
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
			types.ParamsValueCodec(),
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
	for _, listed := range params.LSTAllowlist {
		if listed == bech {
			return true, nil
		}
	}
	return false, nil
}
