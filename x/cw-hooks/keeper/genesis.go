package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v31/x/cw-hooks/types"
)

// InitGenesis import module genesis
func (k Keeper) InitGenesis(
	ctx sdk.Context,
	data types.GenesisState,
) {
	if err := types.ValidateGenesis(data); err != nil {
		panic(err)
	}

	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	for _, v := range data.StakingContractAddresses {
		_, err := sdk.AccAddressFromBech32(v.ContractAddress)
		if err != nil {
			panic(err)
		}

		err = k.SetContract(ctx, types.StakingPrefixKey, v)
		if err != nil {
			panic(err)
		}
	}

	for _, v := range data.GovContractAddresses {
		_, err := sdk.AccAddressFromBech32(v.ContractAddress)
		if err != nil {
			panic(err)
		}

		err = k.SetContract(ctx, types.GovPrefixKey, v)
		if err != nil {
			panic(err)
		}
	}
}

// ExportGenesis export module state
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	stakingContracts, err := k.GetAllContracts(ctx, types.StakingPrefixKey)
	if err != nil {
		panic(err)
	}
	govContracts, err := k.GetAllContracts(ctx, types.GovPrefixKey)
	if err != nil {
		panic(err)
	}
	return &types.GenesisState{
		Params:                   k.GetParams(ctx),
		StakingContractAddresses: stakingContracts,
		GovContractAddresses:     govContracts,
	}
}
