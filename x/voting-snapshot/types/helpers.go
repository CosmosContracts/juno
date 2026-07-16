package types

// DefaultRetentionWindowHeights — pruning is disabled by default at
// v30 launch. The on-chain history is small at activation and the
// safer failure mode (unbounded growth) is preferable to the alternative
// (a buggy or mis-tuned prune sweep silently dropping snapshots). A
// future governance proposal can flip this to, e.g., ~1 year at 2.5s
// blocks (12_614_400) once operators have observed the growth curve.
const DefaultRetentionWindowHeights uint64 = 0

// DefaultPruneInterval — once per block. Lets governance batch the prune
// sweep on busy chains by lifting this knob via params update. 0 and 1
// are treated as equivalent at the keeper.
const DefaultPruneInterval uint64 = 1

// DefaultParams returns the module's default parameter set.
//
// LST allowlist is empty at v30 launch (no live LST contracts on Juno
// per memory/juno-voting-design.md). Default-deny: an unknown LST
// counts as voting power until governance lists it.
func DefaultParams() Params {
	return Params{
		LstAllowlist:           []string{},
		RetentionWindowHeights: DefaultRetentionWindowHeights,
		PruneInterval:          DefaultPruneInterval,
	}
}

// DefaultGenesis returns the module's genesis-time state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{Params: DefaultParams()}
}
