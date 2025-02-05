package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	appparams "github.com/CosmosContracts/juno/v27/app/params"
	"github.com/CosmosContracts/juno/v27/x/tokenfactory/types"
)

// RandDenomCreationFeeParam returns a random DenomCreationFeeParam
func RandDenomCreationFeeParam(r *rand.Rand) sdk.Coins {
	amount := r.Int63n(10_000_000)
	return sdk.NewCoins(sdk.NewCoin(appparams.BondDenom, sdk.NewInt(amount)))
}

// RandomizedGenState generates a random GenesisState for mint
func RandomizedGenState(simstate *module.SimulationState) {
	tfGenesis := types.DefaultGenesis()

	_, err := simstate.Cdc.MarshalJSON(tfGenesis)
	if err != nil {
		panic(err)
	}

	simstate.GenState[types.ModuleName] = simstate.Cdc.MustMarshalJSON(tfGenesis)
}
