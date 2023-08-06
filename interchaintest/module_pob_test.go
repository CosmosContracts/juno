package interchaintest

import (
	"testing"

	"github.com/skip-mev/pob/tests/integration"
	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/stretchr/testify/suite"
)

var (
	numVals = 4
	numFull = 0
)

func GetInterchainSpecForPOB() *interchaintest.ChainSpec {
	// update the genesis kv for juno
	updatedChainConfig := junoConfig
	updatedChainConfig.ModifyGenesis = cosmos.ModifyGenesis(append(defaultGenesisKV, []cosmos.GenesisKV{
		{
			Key:   "app_state.builder.params.max_bundle_size",
			Value: 3,
		},
		{
			Key:   "app_state.builder.params.reserve_fee.denom",
			Value: "ujuno",
		},
		{
			Key:   "app_state.builder.params.reserve_fee.amount",
			Value: "1",
		},
		{
			Key:   "app_state.builder.params.min_bid_increment.denom",
			Value: "ujuno",
		},
		{
			Key:   "app_state.builder.params.min_bid_increment.amount",
			Value: "1",
		},
	}...))
	
	return &interchaintest.ChainSpec{
		Name:          "juno",
		ChainName:     "juno",
		Version:       junoVersion,
		ChainConfig:   updatedChainConfig,
		NumValidators: &numVals,
		NumFullNodes:  &numFull,
	}

}

func TestJunoPOB(t *testing.T) {
	s := integration.NewPOBIntegrationTestSuiteFromSpec(GetInterchainSpecForPOB())
	s.WithDenom("ujuno")

	suite.Run(t, s)
}
