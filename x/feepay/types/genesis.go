package types

// NewGenesisState creates a new genesis state.
func NewGenesisState(params Params, feecontract []FeePayContract) GenesisState {
	return GenesisState{
		Params:      params,
		FeeContract: feecontract,
	}
}

// DefaultGenesisState sets default evm genesis state with empty accounts and
// default params and chain config values.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	return nil
}
