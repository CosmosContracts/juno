package interchaintest

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	feesharetypes "github.com/CosmosContracts/juno/v15/x/feeshare/types"
	tokenfactorytypes "github.com/CosmosContracts/juno/v15/x/tokenfactory/types" // TODO: fix this so we can store in the DB.

	"github.com/docker/docker/client"
	"github.com/icza/dyno"

	interchaintest "github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	testutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
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
		Type:                   "cosmos",
		Name:                   "juno",
		ChainID:                "juno-2",
		Images:                 []ibc.DockerImage{JunoImage},
		Bin:                    "junod",
		Bech32Prefix:           "juno",
		Denom:                  "ujuno",
		CoinType:               "118",
		GasPrices:              "0ujuno",
		GasAdjustment:          1.8,
		TrustingPeriod:         "112h",
		NoHostMount:            false,
		ModifyGenesis:          nil,
		ConfigFileOverrides:    nil,
		EncodingConfig:         junoEncoding(),
		UsingNewGenesisCommand: true,
	}

	pathJunoGaia        = "juno-gaia"
	genesisWalletAmount = int64(10_000_000)
)

// junoEncoding registers the Juno specific module codecs so that the associated types and msgs
// will be supported when writing to the blocksdb sqlite database.
func junoEncoding() *testutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()

	// register custom types
	wasmtypes.RegisterInterfaces(cfg.InterfaceRegistry)
	feesharetypes.RegisterInterfaces(cfg.InterfaceRegistry)
	tokenfactorytypes.RegisterInterfaces(cfg.InterfaceRegistry)

	//github.com/cosmos/cosmos-sdk/types/module/testutil

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

	// votingPeriod := "10s"
	// maxDepositPeriod := "10s"

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:      "juno",
			ChainName: "juno",
			Version:   junoVersion,
			ChainConfig: ibc.ChainConfig{
				// ModifyGenesis: modifyGenesisShortProposals(votingPeriod, maxDepositPeriod),
				Images: []ibc.DockerImage{
					{
						Repository: junoRepo,
						Version:    junoVersion,
						UidGid:     JunoImage.UidGid,
					},
				},
				GasPrices:              "0ujuno",
				Denom:                  "ujuno",
				UsingNewGenesisCommand: true, // v47
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

// Setup Helpers
func modifyGenesisShortProposals(votingPeriod string, maxDepositPeriod string) func(ibc.ChainConfig, []byte) ([]byte, error) {
	return func(chainConfig ibc.ChainConfig, genbz []byte) ([]byte, error) {
		g := make(map[string]interface{})
		if err := json.Unmarshal(genbz, &g); err != nil {
			return nil, fmt.Errorf("failed to unmarshal genesis file: %w", err)
		}
		if err := dyno.Set(g, votingPeriod, "app_state", "gov", "voting_params", "voting_period"); err != nil {
			return nil, fmt.Errorf("failed to set voting period in genesis json: %w", err)
		}
		if err := dyno.Set(g, maxDepositPeriod, "app_state", "gov", "deposit_params", "max_deposit_period"); err != nil {
			return nil, fmt.Errorf("failed to set voting period in genesis json: %w", err)
		}
		if err := dyno.Set(g, chainConfig.Denom, "app_state", "gov", "deposit_params", "min_deposit", 0, "denom"); err != nil {
			return nil, fmt.Errorf("failed to set voting period in genesis json: %w", err)
		}
		out, err := json.Marshal(g)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal genesis bytes to json: %w", err)
		}
		return out, nil
	}
}
