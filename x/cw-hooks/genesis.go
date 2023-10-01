package cwhooks

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v17/x/cw-hooks/keeper"
	"github.com/CosmosContracts/juno/v17/x/cw-hooks/types"
)

// NewGenesisState - Create a new genesis state
func NewGenesisState(params types.Params) *types.GenesisState {
	return &types.GenesisState{
		Params:                   params,
		StakingContractAddresses: []string{},
		GovContractAddresses:     []string{},
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
	return data.Params.Validate()
}

// InitGenesis import module genesis
func InitGenesis(
	ctx sdk.Context,
	k keeper.Keeper,
	data types.GenesisState,
) {
	if err := ValidateGenesis(data); err != nil {
		panic(err)
	}

	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	for _, v := range data.StakingContractAddresses {
		if _, err := sdk.AccAddressFromBech32(v); err != nil {
			panic(err)
		}

		k.SetStakingContract(ctx, types.Contract{
			ContractAddress: v,
		})
	}

	// TODO: gov
}

// ExportGenesis export module state
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	stakingContracts := k.GetAllStakingContract(ctx)

	return &types.GenesisState{
		Params: k.GetParams(ctx),
		StakingContractAddresses: func() []string {
			addresses := make([]string, len(stakingContracts))
			for i, v := range stakingContracts {
				addresses[i] = v.GetContractAddress()
			}
			return addresses
		}(),
	}
}
