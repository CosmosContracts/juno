package types

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState - Create a new genesis state
func NewGenesisState(params Params, stakingContracts, govContracts []string) *GenesisState {
	return &GenesisState{
		Params:                   params,
		StakingContractAddresses: stakingContracts,
		GovContractAddresses:     govContracts,
	}
}

// DefaultGenesisState - Return a default genesis state
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), []string{}, []string{})
}

// GetGenesisStateFromAppState returns x/auth GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.Codec, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}

func ValidateGenesis(data GenesisState) error {
	for _, v := range data.StakingContractAddresses {
		if _, err := sdk.AccAddressFromBech32(v); err != nil {
			return err
		}
	}

	for _, v := range data.GovContractAddresses {
		if _, err := sdk.AccAddressFromBech32(v); err != nil {
			return err
		}
	}

	return data.Params.Validate()
}
