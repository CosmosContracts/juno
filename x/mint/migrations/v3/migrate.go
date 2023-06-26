package v2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/CosmosContracts/juno/v16/app/params"

	"github.com/CosmosContracts/juno/v16/x/mint/exported"
	"github.com/CosmosContracts/juno/v16/x/mint/types"
)

const (
	ModuleName = "mint"
)

var ParamsKey = []byte{0x01}

// Migrate migrates the x/mint module state from the consensus version 2 to
// version 3. Specifically, it takes the parameters that are currently stored
// and managed by the x/params modules and stores them directly into the x/mint
// module state.
func Migrate(
	_ sdk.Context,
	store sdk.KVStore,
	_ exported.Subspace,
	cdc codec.BinaryCodec,
) error {
	// var currParams types.Params
	// legacySubspace.GetParamSet(ctx, &currParams)

	denom, err := sdk.GetBaseDenom()
	if err != nil {
		denom = appparams.BondDenom
	}

	// https://juno-api.reece.sh/cosmos/mint/v1beta1/params
	currParams := types.Params{
		MintDenom:     denom,
		BlocksPerYear: 5048093,
	}

	if err := currParams.Validate(); err != nil {
		return err
	}

	bz := cdc.MustMarshal(&currParams)
	store.Set(ParamsKey, bz)

	return nil
}
