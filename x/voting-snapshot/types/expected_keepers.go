package types

import (
	"context"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// StakingKeeper is the subset of the SDK staking keeper that
// x/voting-snapshot reads to compute and backfill power snapshots.
//
// Note: we deliberately do NOT use staking's GetDelegatorBonded — it sums
// delegations to validators in every bond status, while TotalBondedTokens
// only counts the bonded pool. The keeper computes bonded-validator-only
// power itself via IterateDelegatorDelegations + GetValidator so the
// numerator and denominator share the same basis.
type StakingKeeper interface {
	TotalBondedTokens(ctx context.Context) (math.Int, error)
	IterateAllDelegations(ctx context.Context, fn func(stakingtypes.Delegation) bool) error
	GetValidatorDelegations(ctx context.Context, valAddr sdk.ValAddress) ([]stakingtypes.Delegation, error)
	IterateDelegatorDelegations(ctx context.Context, delegator sdk.AccAddress, cb func(delegation stakingtypes.Delegation) (stop bool)) error
	GetValidator(ctx context.Context, addr sdk.ValAddress) (stakingtypes.Validator, error)
}
