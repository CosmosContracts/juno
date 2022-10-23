package types

import (
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisValidation(t *testing.T) {
	// Valid state
	genState := NewGenesisState(DefaultParams(), []ExchangeRateTuple{}, []FeederDelegation{}, []MissCounter{}, []AggregateExchangeRatePrevote{}, []AggregateExchangeRateVote{})
	require.NoError(t, ValidateGenesis(genState))

	// Invalid Vote Period
	genState.Params.VotePeriod = 0
	require.Error(t, ValidateGenesis(genState))

	// Invalid VoteThreshold
	genState = DefaultGenesisState()
	genState.Params.VoteThreshold = sdk.NewDecWithPrec(33, 2)
	require.Error(t, ValidateGenesis(genState))

	// Invalid Rewardband
	genState = DefaultGenesisState()
	genState.Params.RewardBand = sdk.NewDec(2)
	require.Error(t, ValidateGenesis(genState))
	genState.Params.RewardBand = sdk.NewDec(-1)
	require.Error(t, ValidateGenesis(genState))

	// Invalid RewardDistributionWindow
	genState = DefaultGenesisState()
	genState.Params.RewardDistributionWindow = genState.Params.VotePeriod - 1
	require.Error(t, ValidateGenesis(genState))

	// Invalid SlashFraction
	genState = DefaultGenesisState()
	genState.Params.SlashFraction = sdk.NewDec(2)
	require.Error(t, ValidateGenesis(genState))
	genState.Params.SlashFraction = sdk.NewDec(-1)
	require.Error(t, ValidateGenesis(genState))

	// Invalid SlashWindow
	genState = DefaultGenesisState()
	genState.Params.SlashWindow = genState.Params.VotePeriod - 1
	require.Error(t, ValidateGenesis(genState))

	// Invalid MinValidPerWindow
	genState = DefaultGenesisState()
	genState.Params.MinValidPerWindow = sdk.NewDec(2)
	require.Error(t, ValidateGenesis(genState))
	genState.Params.MinValidPerWindow = sdk.NewDec(-1)
	require.Error(t, ValidateGenesis(genState))

	// Invalid AcceptList
	genState = DefaultGenesisState()
	genState.Params.AcceptList = DenomList{Denom{}}
	require.Error(t, ValidateGenesis(genState))
}

func TestGetGenesisStateFromAppState(t *testing.T) {
	emptyGenesis := GenesisState{
		Params:                        Params{},
		ExchangeRates:                 []ExchangeRateTuple{},
		FeederDelegations:             []FeederDelegation{},
		MissCounters:                  []MissCounter{},
		AggregateExchangeRatePrevotes: []AggregateExchangeRatePrevote{},
		AggregateExchangeRateVotes:    []AggregateExchangeRateVote{},
	}

	bz, err := json.Marshal(emptyGenesis)
	require.Nil(t, err)

	require.NotNil(t, GetGenesisStateFromAppState(ModuleCdc, map[string]json.RawMessage{
		ModuleName: bz,
	}))
	require.NotNil(t, GetGenesisStateFromAppState(ModuleCdc, map[string]json.RawMessage{}))
}
