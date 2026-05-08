package types

import "encoding/json"

// DefaultRetentionWindowHeights — roughly one year at 2.5s blocks.
// Old snapshots beyond this window are pruned by the module's
// EndBlocker. The value is intentionally conservative; governance
// can shorten it once historical-vote consumers settle on a tighter
// proposal-window upper bound. Set to 0 to disable pruning.
const DefaultRetentionWindowHeights uint64 = 12_614_400

// Params is the module's governance-controlled configuration.
//
// LSTAllowlist holds the bech32 addresses of LST contracts whose
// delegations must NOT count toward voting power. Per Juno's voting
// design (memory/juno-voting-design.md), LSTs do not contribute voting
// power because the wrapper contract — not the underlying staker —
// holds the delegation. The list is empty at v30 launch (no live LST
// contracts on Juno). Default-deny: an unknown LST counts as voting
// power until governance lists it.
//
// RetentionWindowHeights bounds how far back (delegator, height) and
// (height) snapshots are retained. Snapshots with height <
// (current_height - RetentionWindowHeights) are pruned by the
// EndBlocker. Zero disables pruning.
type Params struct {
	LSTAllowlist           []string `json:"lst_allowlist"`
	RetentionWindowHeights uint64   `json:"retention_window_heights"`
}

func DefaultParams() Params {
	return Params{
		LSTAllowlist:           []string{},
		RetentionWindowHeights: DefaultRetentionWindowHeights,
	}
}

func (p Params) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func (p *Params) Unmarshal(b []byte) error {
	return json.Unmarshal(b, p)
}
