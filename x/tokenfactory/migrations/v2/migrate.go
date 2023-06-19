package v2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v16/x/tokenfactory/exported"
	"github.com/CosmosContracts/juno/v16/x/tokenfactory/types"
)

const ModuleName = "tokenfactory"

var ParamsKey = []byte{0x00}

// Migrate migrates the x/tokenfactory module state from the consensus version 1 to
// version 2. Specifically, it takes the parameters that are currently stored
// and managed by the x/params modules and stores them directly into the x/tokenfactory
// module state.
func Migrate(
	ctx sdk.Context,
	store sdk.KVStore,
	legacySubspace exported.Subspace,
	cdc codec.BinaryCodec,
) error {
	// TODO: The following breaks for all modules for some reason except FeeShare.
	// These params matche our mainnet.

	// var currParams types.Params
	// legacySubspace.GetParamSet(ctx, &currParams)

	currParams := types.Params{
		DenomCreationFee:        nil,
		DenomCreationGasConsume: 2_000_000,
	}

	if err := currParams.Validate(); err != nil {
		return err
	}

	bz := cdc.MustMarshal(&currParams)
	store.Set(ParamsKey, bz)

	return nil
}