package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/CosmosContracts/juno/v7/x/mint/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

const (
	keyBlocksPerYear = "BlocksPerYear"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, keyBlocksPerYear,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenBlocksPerYear(r))
			},
		),
	}
}
