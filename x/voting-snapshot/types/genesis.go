package types

import "encoding/json"

type GenesisState struct {
	Params Params `json:"params"`
}

func DefaultGenesis() *GenesisState {
	return &GenesisState{Params: DefaultParams()}
}

func (g GenesisState) Marshal() ([]byte, error)  { return json.Marshal(g) }
func (g *GenesisState) Unmarshal(b []byte) error { return json.Unmarshal(b, g) }
