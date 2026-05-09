package types

// DefaultRetentionWindowHeights — roughly one year at 2.5s blocks.
// Old snapshots beyond this window are pruned by the module's
// EndBlocker. Set to 0 to disable pruning.
const DefaultRetentionWindowHeights uint64 = 12_614_400

// DefaultParams returns the module's default parameter set.
//
// LST allowlist is empty at v30 launch (no live LST contracts on Juno
// per memory/juno-voting-design.md). Default-deny: an unknown LST
// counts as voting power until governance lists it.
func DefaultParams() Params {
	return Params{
		LstAllowlist:           []string{},
		RetentionWindowHeights: DefaultRetentionWindowHeights,
	}
}

// DefaultGenesis returns the module's genesis-time state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{Params: DefaultParams()}
}
