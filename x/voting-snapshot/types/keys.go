package types

import "cosmossdk.io/collections"

const (
	ModuleName = "votingsnapshot"
	StoreKey   = ModuleName

	// TransientStoreKey backs the per-block dirty-delegator set. Staking
	// hooks mark delegators dirty here; the module EndBlocker drains the
	// set into persistent snapshots and the store auto-resets on commit.
	TransientStoreKey = "transient_votingsnapshot"
)

// persistent store prefixes
var (
	ParamsKey      = collections.NewPrefix(0)
	VotingPowerKey = collections.NewPrefix(1) // (delegator, height) -> sdkmath.Int
	TotalPowerKey  = collections.NewPrefix(2) // height           -> sdkmath.Int
)

// transient store prefixes (separate store — no clash with the above)
var (
	TransientDirtyDelegatorsKey = collections.NewPrefix(0) // delegator -> ()
	TransientTotalDirtyKey      = collections.NewPrefix(1) // () -> bool
)
