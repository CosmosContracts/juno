package simulation

import (
	"fmt"
	"math/rand"

	"github.com/CosmosContracts/juno/v15/x/tokenfactory/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

func ParamChanges(_ *rand.Rand) []simtypes.LegacyParamChange {
	return []simtypes.LegacyParamChange{
		simulation.NewSimLegacyParamChange(
			types.ModuleName,
			string(types.KeyDenomCreationFee),
			func(r *rand.Rand) string {
				amount := RandDenomCreationFeeParam(r)
				return fmt.Sprintf("[{\"denom\":\"%v\",\"amount\":\"%v\"}]", amount[0].Denom, amount[0].Amount)
			},
		),
	}
}
