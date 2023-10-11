package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// NewGenesisState creates a new genesis state.
func NewGenesisState(params Params, feePayContracts []FeePayContract) GenesisState {
	return GenesisState{
		Params:          params,
		FeePayContracts: feePayContracts,
	}
}

// DefaultGenesisState sets default genesis state with empty accounts and
// default params and chain config values.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: Params{
			EnableFeepay: true,
		},
		FeePayContracts: []FeePayContract{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Loop through all fee pay contracts and validate they
	// have a valid bech32 address
	for _, contract := range gs.FeePayContracts {
		if _, err := sdk.AccAddressFromBech32(contract.ContractAddress); err != nil {
			return err
		}
	}

	return nil
}
