package types

import "encoding/json"

// Params is the module's governance-controlled configuration.
//
// LSTAllowlist holds the bech32 addresses of LST contracts whose
// delegations must NOT count toward voting power. Per Juno's voting
// design (memory/juno-voting-design.md), LSTs do not contribute voting
// power because the wrapper contract — not the underlying staker —
// holds the delegation. The list is empty at v30 launch (no live LST
// contracts on Juno). Default-deny: an unknown LST counts as voting
// power until governance lists it.
type Params struct {
	LSTAllowlist []string `json:"lst_allowlist"`
}

func DefaultParams() Params {
	return Params{LSTAllowlist: []string{}}
}

func (p Params) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func (p *Params) Unmarshal(b []byte) error {
	return json.Unmarshal(b, p)
}
