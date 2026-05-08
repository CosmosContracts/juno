package types

import "cosmossdk.io/collections"

const (
	ModuleName = "votingsnapshot"
	StoreKey   = ModuleName
)

var (
	ParamsKey      = collections.NewPrefix(0)
	VotingPowerKey = collections.NewPrefix(1) // (delegator, height) -> sdkmath.Int
	TotalPowerKey  = collections.NewPrefix(2) // height           -> sdkmath.Int
)
