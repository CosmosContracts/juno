package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"

	sdktestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/icza/dyno"
	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	interchaintestrelayer "github.com/strangelove-ventures/interchaintest/v7/relayer"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	sdk "github.com/cosmos/cosmos-sdk/types"

	// params
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	feesharetypes "github.com/CosmosContracts/juno/v15/x/feeshare/types"
	tokenfactorytypes "github.com/CosmosContracts/juno/v15/x/tokenfactory/types"
)

func junoEncoding() *sdktestutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()

	// register custom types
	wasmtypes.RegisterInterfaces(cfg.InterfaceRegistry)
	feesharetypes.RegisterInterfaces(cfg.InterfaceRegistry)
	tokenfactorytypes.RegisterInterfaces(cfg.InterfaceRegistry)

	return &cfg
}

func modifyGenesis(genesis Genesis) func(ibc.ChainConfig, []byte) ([]byte, error) {
	return func(chainConfig ibc.ChainConfig, genbz []byte) ([]byte, error) {
		g := make(map[string]interface{})
		if err := json.Unmarshal(genbz, &g); err != nil {
			return nil, fmt.Errorf("failed to unmarshal genesis file: %w", err)
		}

		for idx, values := range genesis.Modify {
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

	// ibc-path-name -> index of []cosmos.CosmosChain
	ibcpaths := make(map[string][]int)
	chainSpecs := []*interchaintest.ChainSpec{}

	for idx, cfg := range config.Chains {
		if cfg.Debugging {
			t.Logf("[%d] %v", idx, cfg)
		}

		chainConfig := ibc.ChainConfig{
			Type:                cfg.ChainType,
			Name:                cfg.Name,
			ChainID:             cfg.ChainID,
			Bin:                 cfg.Binary,
			Bech32Prefix:        cfg.Bech32Prefix,
			Denom:               cfg.Denom,
			CoinType:            fmt.Sprintf("%d", cfg.CoinType),
			GasPrices:           cfg.GasPrices,
			GasAdjustment:       cfg.GasAdjustment,
			TrustingPeriod:      cfg.TrustingPeriod,
			NoHostMount:         false,
			ModifyGenesis:       modifyGenesis(cfg.Genesis),
			ConfigFileOverrides: nil,
			EncodingConfig:      junoEncoding(),
		}

		chainConfig.Images = []ibc.DockerImage{{
			Repository: cfg.DockerImage.Repository,
			Version:    cfg.DockerImage.Version,
			UidGid:     cfg.DockerImage.UidGid,
		}}

		chainSpecs = append(chainSpecs, &interchaintest.ChainSpec{
			Name:          cfg.Name,
			Version:       cfg.DockerImage.Version,
			ChainName:     cfg.ChainID,
			ChainConfig:   chainConfig,
			NumValidators: &cfg.NumberVals,
			NumFullNodes:  &cfg.NumberNode,
		})

		if cfg.IBCPath != "" {
			fmt.Println("IBC Path:", cfg.IBCPath, "Chain:", cfg.Name)
			ibcpaths[cfg.IBCPath] = append(ibcpaths[cfg.IBCPath], idx)
		}
	}

	// ensure that none of ibcpaths values are length > 2
	for k, v := range ibcpaths {
		if len(v) == 1 {
			t.Fatalf("ibc path '%s' has only 1 chain", k)
		}
		if len(v) > 2 {
			t.Fatalf("ibc path '%s' has more than 2 chains", k)
		}
	}

	// Create chain factory for all the chains
	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), chainSpecs)

	// Get chains from the chain factory
	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	// iterate all chains chain's configs & setup accounts
	additionalWallets := make(map[ibc.Chain][]ibc.WalletAmount)
	for idx, chain := range config.Chains {
		chainObj := chains[idx].(*cosmos.CosmosChain)

		for _, acc := range chain.Genesis.Accounts {
			amount, err := sdk.ParseCoinsNormalized(acc.Amount)
			if err != nil {
				panic(err)
			}

			for _, coin := range amount {
				additionalWallets[chainObj] = append(additionalWallets[chainObj], ibc.WalletAmount{
					Address: acc.Address,
					Amount:  coin.Amount.Int64(),
					Denom:   coin.Denom,
				})
			}
		}
	}

	// Create a new Interchain object which describes the chains, relayers, and IBC connections we want to use
	ic := interchaintest.NewInterchain()
	for _, chain := range chains {
		// fmt.Println("adding chain...", chain.Config().Name)
		ic = ic.AddChain(chain)
	}
	ic.AdditionalGenesisWallets = additionalWallets

	// Base setup
	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)
	ctx := context.Background()
	client, network := interchaintest.DockerSetup(t)

	// setup a relayer if we have IBC paths to use
	if len(ibcpaths) > 0 {
		// relayer
		// Get a relayer instance
		relayerType, relayerName := ibc.CosmosRly, "relay"
		rf := interchaintest.NewBuiltinRelayerFactory(
			relayerType,
			zaptest.NewLogger(t),
			// TODO: put into the config
			interchaintestrelayer.CustomDockerImage("ghcr.io/cosmos/relayer", "latest", "100:1000"),
			interchaintestrelayer.StartupFlags("--processor", "events", "--block-history", "100"),
		)

		r := rf.Build(t, client, network)

		ic = ic.AddRelayer(r, relayerName)

		// Add links between chains
		for path, c := range ibcpaths {
			interLink := interchaintest.InterchainLink{
				Chain1:  nil,
				Chain2:  nil,
				Path:    path,
				Relayer: r,
			}

			// set chain1 & chain2
			for idx, chain := range c {
				if idx == 0 {
					interLink.Chain1 = chains[chain]
				} else {
					interLink.Chain2 = chains[chain]
				}
			}

			fmt.Print(interLink)
			ic = ic.AddLink(interLink)
		}
	}

	// Build all chains & begin.
	err = ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: false,
		// BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),
	})
	require.NoError(t, err)

	// wait for blocks
	var outputLogs []LogOutput
	var longestTTLChain *cosmos.CosmosChain
	ttlWait := 0
	for idx, chain := range config.Chains {
		chainObj := chains[idx].(*cosmos.CosmosChain)
		t.Logf("\n\n\n\nWaiting for %d blocks on chain %s", chain.BlocksTTL, chainObj.Config().ChainID)

		v := LogOutput{
			// TODO: Rest Address?
			ChainID:     chainObj.Config().ChainID,
			ChainName:   chainObj.Config().Name,
			RPCAddress:  chainObj.GetHostRPCAddress(),
			GRPCAddress: chainObj.GetHostGRPCAddress(),
			IBCPath:     chain.IBCPath,
		}

		if chain.BlocksTTL > ttlWait {
			ttlWait = chain.BlocksTTL
			longestTTLChain = chainObj
		}

		outputLogs = append(outputLogs, v)
	}

	// dump output logs to file
	bz, _ := json.MarshalIndent(outputLogs, "", "  ")
	ioutil.WriteFile("logs.json", []byte(bz), 0644)

	// TODO: Way for us to wait for blocks & show the tx logs during this time for each block?
	if err = testutil.WaitForBlocks(ctx, ttlWait, longestTTLChain); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = ic.Close()
		// TODO: also delete logs.json file? or a file which is tmnp
	})
}
