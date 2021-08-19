package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/CosmosContracts/juno/x/mint/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

const (
<<<<<<< HEAD
	keyInflationRateChange = "InflationRateChange"
	keyInflationMax        = "InflationMax"
	keyInflationMin        = "InflationMin"
	keyGoalBonded          = "GoalBonded"
=======
	keyBlocksPerYear = "BlocksPerYear"
>>>>>>> disperze/mint-module
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
<<<<<<< HEAD
		simulation.NewSimParamChange(types.ModuleName, keyInflationRateChange,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenInflationRateChange(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyInflationMax,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenInflationMax(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyInflationMin,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenInflationMin(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyGoalBonded,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenGoalBonded(r))
=======
		simulation.NewSimParamChange(types.ModuleName, keyBlocksPerYear,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenBlocksPerYear(r))
>>>>>>> disperze/mint-module
			},
		),
	}
}
