package cwhooks

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v22/x/cw-hooks/keeper"
	"github.com/CosmosContracts/juno/v22/x/cw-hooks/types"
)

// NewGenesisState - Create a new genesis state
func NewGenesisState(params types.Params, stakingContracts, govContracts []string) *types.GenesisState {
	return &types.GenesisState{
		Params:                   params,
		StakingContractAddresses: stakingContracts,
		GovContractAddresses:     govContracts,
	}
}

// DefaultGenesisState - Return a default genesis state
func DefaultGenesisState() *types.GenesisState {
	return NewGenesisState(types.DefaultParams(), []string{}, []string{})
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
		accAddr, err := sdk.AccAddressFromBech32(v)
		if err != nil {
			panic(err)
		}

		k.SetContract(ctx, types.KeyPrefixStaking, accAddr)
	}

	for _, v := range data.GovContractAddresses {
		accAddr, err := sdk.AccAddressFromBech32(v)
		if err != nil {
			panic(err)
		}

		k.SetContract(ctx, types.KeyPrefixGov, accAddr)
	}
}

// ExportGenesis export module state
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:                   k.GetParams(ctx),
		StakingContractAddresses: k.GetAllContractsBech32(ctx, types.KeyPrefixStaking),
		GovContractAddresses:     k.GetAllContractsBech32(ctx, types.KeyPrefixGov),
	}
}
