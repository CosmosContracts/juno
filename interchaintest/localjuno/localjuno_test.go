package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/icza/dyno"
	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	// params
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	feesharetypes "github.com/CosmosContracts/juno/v15/x/feeshare/types"
	tokenfactorytypes "github.com/CosmosContracts/juno/v15/x/tokenfactory/types"
)

func junoEncoding() *testutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()

	// register custom types
	wasmtypes.RegisterInterfaces(cfg.InterfaceRegistry)
	feesharetypes.RegisterInterfaces(cfg.InterfaceRegistry)
	tokenfactorytypes.RegisterInterfaces(cfg.InterfaceRegistry)

	return &cfg
}

func modifyGenesis(cfg *MainConfig) func(ibc.ChainConfig, []byte) ([]byte, error) {
	return func(chainConfig ibc.ChainConfig, genbz []byte) ([]byte, error) {
		g := make(map[string]interface{})
		if err := json.Unmarshal(genbz, &g); err != nil {
			return nil, fmt.Errorf("failed to unmarshal genesis file: %w", err)
		}

		for idx, values := range cfg.Chains[0].Genesis.Modify {
			path := strings.Split(values.Key, ".")

			result := make([]interface{}, len(path))
			for i, component := range path {
				if v, err := strconv.Atoi(component); err == nil {
					result[i] = v
				} else {
					result[i] = component
				}
			}

			if err := dyno.Set(g, values.Value, result...); err != nil {
				return nil, fmt.Errorf("failed to set value (index:%d) in genesis json: %w", idx, err)
			}
		}

		out, err := json.Marshal(g)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal genesis bytes to json: %w", err)
		}
		return out, nil
	}
}

// TestLocalJuno runs a local juno chain easily.
func TestLocalJuno(t *testing.T) {
	config, err := LoadConfig()
	require.NoError(t, err)

	cfgJuno := config.Chains[0]

	chainConfig := ibc.ChainConfig{
		Type:                "cosmos",
		Name:                "juno",
		ChainID:             cfgJuno.ChainID,
		Bin:                 "junod",
		Bech32Prefix:        "juno",
		Denom:               "ujuno",
		CoinType:            "118",
		GasPrices:           cfgJuno.GasPrices,
		GasAdjustment:       cfgJuno.GasAdjustment,
		TrustingPeriod:      "112h",
		NoHostMount:         false,
		ModifyGenesis:       modifyGenesis(config),
		ConfigFileOverrides: nil,
		EncodingConfig:      junoEncoding(),
	}

	if cfgJuno.Version.Type == "branch" {
		chainConfig.Images = []ibc.DockerImage{{
			Repository: "ghcr.io/cosmoscontracts/juno-e2e",
			Version:    cfgJuno.Version.Version,
			UidGid:     "1025:1025",
		}}
	}

	// Create chain factory with Juno
	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:          "juno",
			Version:       cfgJuno.Version.Version,
			ChainName:     cfgJuno.ChainID,
			ChainConfig:   chainConfig,
			NumValidators: &cfgJuno.NumberVals,
			NumFullNodes:  &cfgJuno.NumberNode,
		},
	})

	// Get chains from the chain factory
	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	juno := chains[0].(*cosmos.CosmosChain)

	// Create a new Interchain object which describes the chains, relayers, and IBC connections we want to use
	ic := interchaintest.NewInterchain().AddChain(juno)

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	ctx := context.Background()
	client, network := interchaintest.DockerSetup(t)

	err = ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:          t.Name(),
		Client:            client,
		NetworkID:         network,
		SkipPathCreation:  true,
		BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = ic.Close()
	})
}

// home, err := os.UserHomeDir()
// if err != nil {
// 	panic(err)
// }
// return filepath.Join(home, ".ibctest", "databases", "block.db")
