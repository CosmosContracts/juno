package v2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/CosmosContracts/juno/v16/app/params"
	"github.com/CosmosContracts/juno/v16/x/globalfee/keeper/exported"
	"github.com/CosmosContracts/juno/v16/x/globalfee/types"
)

const (
	ModuleName = "globalfee"
)

var ParamsKey = []byte{0x00}

// Migrate migrates the x/globalfee module state from the consensus version 1 to
// version 2. Specifically, it takes the parameters that are currently stored
// and managed by the x/params modules and stores them directly into the x/globalfee
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

	// https://juno-api.reece.sh/gaia/globalfee/v1beta1/minimum_gas_prices
	currParams := types.Params{
		MinimumGasPrices: sdk.DecCoins{
			// 0.003000000000000000uatom
			sdk.NewDecCoinFromDec("ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9", sdk.NewDecWithPrec(3, 3)),
			// 0.075000000000000000ujuno / ujunox
			sdk.NewDecCoinFromDec(denom, sdk.NewDecWithPrec(75, 3)),
		},
	}

	if err := currParams.Validate(); err != nil {
		return err
	}
	bz := cdc.MustMarshal(&currParams)
	store.Set(ParamsKey, bz)

	return nil
}
