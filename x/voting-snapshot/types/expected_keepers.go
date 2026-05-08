package types

import (
	"context"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// StakingKeeper is the subset of the SDK staking keeper that
// x/voting-snapshot reads to compute and backfill power snapshots.
type StakingKeeper interface {
	GetDelegatorBonded(ctx context.Context, delegator sdk.AccAddress) (math.Int, error)
	TotalBondedTokens(ctx context.Context) (math.Int, error)
	IterateAllDelegations(ctx context.Context, fn func(stakingtypes.Delegation) bool) error
	GetValidatorDelegations(ctx context.Context, valAddr sdk.ValAddress) ([]stakingtypes.Delegation, error)
}
