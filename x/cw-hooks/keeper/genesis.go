package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v27/x/cw-hooks/types"
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
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params:                   k.GetParams(ctx),
		StakingContractAddresses: k.GetAllContractsBech32(ctx, types.KeyPrefixStaking),
		GovContractAddresses:     k.GetAllContractsBech32(ctx, types.KeyPrefixGov),
	}
}
