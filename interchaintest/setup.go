package interchaintest

import (
	"context"
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	feesharetypes "github.com/CosmosContracts/juno/v14/x/feeshare/types"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/docker/docker/client"

	interchaintest "github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
	"github.com/strangelove-ventures/interchaintest/v4/testreporter"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

var (
	JunoE2ERepo  = "ghcr.io/cosmoscontracts/juno-e2e"
	JunoMainRepo = "ghcr.io/cosmoscontracts/juno"

	junoRepo, junoVersion = GetDockerImageInfo()

	JunoImage = ibc.DockerImage{
		Repository: junoRepo,
		Version:    junoVersion,
		UidGid:     "1025:1025",
	}

	junoConfig = ibc.ChainConfig{
		Type:                "cosmos",
		Name:                "juno",
		ChainID:             "juno-2",
		Images:              []ibc.DockerImage{JunoImage},
		Bin:                 "junod",
		Bech32Prefix:        "juno",
		Denom:               "ujuno",
		CoinType:            "118",
		GasPrices:           "0.0ujuno",
		GasAdjustment:       1.1,
		TrustingPeriod:      "112h",
		NoHostMount:         false,
		SkipGenTx:           false,
		PreGenesis:          nil,
		ModifyGenesis:       nil,
		ConfigFileOverrides: nil,
		EncodingConfig:      junoEncoding(),
	}

	pathJunoGaia        = "juno-gaia"
	genesisWalletAmount = int64(10_000_000)
)

// junoEncoding registers the Juno specific module codecs so that the associated types and msgs
// will be supported when writing to the blocksdb sqlite database.
func junoEncoding() *simappparams.EncodingConfig {
	cfg := cosmos.DefaultEncoding()

	// register custom types
	feesharetypes.RegisterInterfaces(cfg.InterfaceRegistry)
	wasmtypes.RegisterInterfaces(cfg.InterfaceRegistry)

	return &cfg
}

// Basic chain setup for a Juno chain. No relaying
func CreateBaseChain(t *testing.T) []ibc.Chain {
	// Create chain factory with Juno
	numVals := 1
	numFullNodes := 0

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:      "juno",
			Version:   "latest",
			ChainName: "juno1",
			ChainConfig: ibc.ChainConfig{
				GasPrices:      "0ujuno",
				GasAdjustment:  2.0,
				EncodingConfig: junoEncoding(),
			},
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
	})

	// Get chains from the chain factory
	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	// juno := chains[0].(*cosmos.CosmosChain)
	return chains
}

func CreateThisBranchChain(t *testing.T) []ibc.Chain {
	// Create chain factory with Juno
	numVals := 1
	numFullNodes := 0

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:      "juno",
			ChainName: "juno",
			Version:   junoVersion,
			ChainConfig: ibc.ChainConfig{
				// ModifyGenesis: cosmos.ModifyGenesisProposalTime(votingPeriod, maxDepositPeriod),
				Images: []ibc.DockerImage{
					{
						Repository: junoRepo,
						Version:    junoVersion,
						UidGid:     JunoImage.UidGid,
					},
				},
			},
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
	})

	// Get chains from the chain factory
	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	// juno := chains[0].(*cosmos.CosmosChain)
	return chains
}

func BuildInitialChain(t *testing.T, chains []ibc.Chain) (*interchaintest.Interchain, context.Context, *client.Client, string) {
	// Create a new Interchain object which describes the chains, relayers, and IBC connections we want to use
	ic := interchaintest.NewInterchain()

	for _, chain := range chains {
		ic.AddChain(chain)
	}

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	ctx := context.Background()
	client, network := interchaintest.DockerSetup(t)

	err := ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: true,
		// This can be used to write to the block database which will index all block data e.g. txs, msgs, events, etc.
		// BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),
	})
	require.NoError(t, err)

	return ic, ctx, client, network
}
