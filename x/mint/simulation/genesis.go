package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/CosmosContracts/juno/v27/x/mint/types"
)

// Simulation parameter constants
const (
	Inflation = "inflation"
)

// GenInflation randomized Inflation
func GenInflation(r *rand.Rand) sdkmath.LegacyDec {
	return sdkmath.LegacyNewDecWithPrec(int64(r.Intn(99)), 2)
}

// GenBlocksPerYear randomized BlocksPerYear
func GenBlocksPerYear(_ *rand.Rand) uint64 {
	return uint64(60 * 60 * 8766 / 5)
}

// RandomizedGenState generates a random GenesisState for mint
func RandomizedGenState(simState *module.SimulationState) {
	// minter
	var inflation sdkmath.LegacyDec
	simState.AppParams.GetOrGenerate(
		Inflation, &inflation, simState.Rand,
		func(r *rand.Rand) { inflation = GenInflation(r) },
	)

	// params
	mintDenom := sdk.DefaultBondDenom
	blocksPerYear := uint64(60 * 60 * 8766 / 5)
	params := types.NewParams(mintDenom, blocksPerYear)

	mintGenesis := types.NewGenesisState(types.InitialMinter(inflation), params)

	bz, err := json.MarshalIndent(&mintGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated minting parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(mintGenesis)
}
