package interchaintest

import (
	"context"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/conformance"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TestJunoGaiaIBCTransfer spins up a Juno and Gaia network, initializes an IBC connection between them,
// and sends an ICS20 token transfer from Juno->Gaia and then back from Gaia->Juno.
func TestJunoGaiaIBCTransfer(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	// Create chain factory with Juno and Gaia
	numVals := 1
	numFullNodes := 1

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:          "juno",
			ChainConfig:   junoConfig,
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
		// {
		// 	Name: "juno",
		// 	ChainConfig: ibc.ChainConfig{
		// 		Type:                   "cosmos",
		// 		Name:                   "juno",
		// 		ChainID:                "juno-3",
		// 		Images:                 []ibc.DockerImage{JunoImage},
		// 		Bin:                    "junod",
		// 		Bech32Prefix:           "juno",
		// 		Denom:                  "ujuno",
		// 		CoinType:               "118",
		// 		GasPrices:              "0ujuno",
		// 		GasAdjustment:          2.0,
		// 		TrustingPeriod:         "112h",
		// 		NoHostMount:            false,
		// 		ConfigFileOverrides:    nil,
		// 		EncodingConfig:         junoEncoding(),
		// 		UsingNewGenesisCommand: true,
		// 		ModifyGenesis:          modifyGenesisShortProposals(VotingPeriod, MaxDepositPeriod),
		// 	},
		// 	NumValidators: &numVals,
		// 	NumFullNodes:  &numFullNodes,
		// },
		{
			Name:          "gaia",
			Version:       "v9.0.0",
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
	})

	const (
		path        = "ibc-path"
		relayerName = "relayer"
	)

	// Get chains from the chain factory
	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	client, network := interchaintest.DockerSetup(t)

	juno, gaia := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)

	// Get a relayer instance
	rf := interchaintest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		zaptest.NewLogger(t),
	)

	r := rf.Build(t, client, network)

	ic := interchaintest.NewInterchain().
		AddChain(juno).
		AddChain(gaia).
		AddRelayer(r, relayerName).
		AddLink(interchaintest.InterchainLink{
			Chain1:  juno,
			Chain2:  gaia,
			Relayer: r,
			Path:    path,
		})

	ctx := context.Background()

	rep := testreporter.NewNopReporter()

	require.NoError(t, ic.Build(ctx, rep.RelayerExecReporter(t), interchaintest.InterchainBuildOptions{
		TestName:          t.Name(),
		Client:            client,
		NetworkID:         network,
		BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),
		SkipPathCreation:  false,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	// test IBC conformance
	conformance.TestChainPair(t, ctx, client, network, juno, gaia, rf, rep, r, path)
}
