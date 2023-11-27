package clock

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v18/x/clock/keeper"
	"github.com/CosmosContracts/juno/v18/x/clock/types"
)

// NewGenesisState - Create a new genesis state
func NewGenesisState(params types.Params) *types.GenesisState {
	return &types.GenesisState{
		Params: params,
	}
}

// DefaultGenesisState - Return a default genesis state
func DefaultGenesisState() *types.GenesisState {
	return NewGenesisState(types.DefaultParams())
}

// GetGenesisStateFromAppState returns x/auth GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.Codec, appState map[string]json.RawMessage) *types.GenesisState {
	var genesisState types.GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}

func ValidateGenesis(data types.GenesisState) error {
	err := data.Params.Validate()
	if err != nil {
		return err
	}

	// Validate contracts
	for _, addr := range data.ContractAddresses {
		if _, err := sdk.AccAddressFromBech32(addr); err != nil {
			return err
		}
	}

	// Validate jailed contracts
	for _, addr := range data.JailedContractAddresses {
		if _, err := sdk.AccAddressFromBech32(addr); err != nil {
			return err
		}
	}

	return nil
}

// InitGenesis import module genesis
func InitGenesis(
	ctx sdk.Context,
	k keeper.Keeper,
	data types.GenesisState,
) {
	// Validate init contents
	if err := ValidateGenesis(data); err != nil {
		panic(err)
	}

	// Set params
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	// Register unjailed contracts
	for _, addr := range data.ContractAddresses {
		k.SetClockContract(ctx, addr, false)
	}

	// Register jailed contracts
	for _, addr := range data.JailedContractAddresses {
		k.SetClockContract(ctx, addr, true)
	}
}

// ExportGenesis export module state
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	params := k.GetParams(ctx)
	contracts := k.GetAllContracts(ctx, false)
	jailedContracts := k.GetAllContracts(ctx, true)

	return &types.GenesisState{
		Params:                  params,
		ContractAddresses:       contracts,
		JailedContractAddresses: jailedContracts,
	}
}
