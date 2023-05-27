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

	cfgJuno := config.Chains[0]

	fmt.Println(cfgJuno)

	chainConfig := ibc.ChainConfig{
		Type:                cfgJuno.ChainType,
		Name:                cfgJuno.Name,
		ChainID:             cfgJuno.ChainID,
		Bin:                 cfgJuno.Binary,
		Bech32Prefix:        cfgJuno.Bech32Prefix,
		Denom:               cfgJuno.Denom,
		CoinType:            fmt.Sprintf("%d", cfgJuno.CoinType),
		GasPrices:           cfgJuno.GasPrices,
		GasAdjustment:       cfgJuno.GasAdjustment,
		TrustingPeriod:      cfgJuno.TrustingPeriod,
		NoHostMount:         false,
		ModifyGenesis:       modifyGenesis(cfgJuno.Genesis),
		ConfigFileOverrides: nil,
		EncodingConfig:      junoEncoding(),
	}

	chainConfig.Images = []ibc.DockerImage{{
		Repository: cfgJuno.DockerImage.Repository,
		Version:    cfgJuno.DockerImage.Version,
		UidGid:     cfgJuno.DockerImage.UidGid,
	}}

	// Create chain factory with Juno
	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:          cfgJuno.Name,
			Version:       cfgJuno.DockerImage.Version,
			ChainName:     cfgJuno.ChainID,
			ChainConfig:   chainConfig,
			NumValidators: &cfgJuno.NumberVals,
			NumFullNodes:  &cfgJuno.NumberNode,
		},
	})

	// Get chains from the chain factory
	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	// iterate all chains config.Chains & setup accounts
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

	juno := chains[0].(*cosmos.CosmosChain)

	// Create a new Interchain object which describes the chains, relayers, and IBC connections we want to use
	ic := interchaintest.NewInterchain().AddChain(juno)
	ic.AdditionalGenesisWallets = additionalWallets

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

	// wait for blocks
	for idx, chain := range config.Chains {
		chainObj := chains[idx].(*cosmos.CosmosChain)
		t.Logf("\n\n\n\nWaiting for %d blocks on chain %s", chain.BlocksTTL, chainObj.Config().ChainID)

		v := LogOutput{
			// TODO: Rest Address?
			ChainID:     chainObj.Config().ChainID,
			ChainName:   chainObj.Config().Name,
			RPCAddress:  chainObj.GetHostRPCAddress(),
			GRPCAddress: chainObj.GetHostGRPCAddress(),
		}
		bz, _ := json.MarshalIndent(v, "", "  ")
		ioutil.WriteFile("logs.json", []byte(bz), 0644)

		if err = testutil.WaitForBlocks(ctx, chain.BlocksTTL, chainObj); err != nil {
			// TODO: Way for us to wait for blocks & show the tx logs?
			t.Fatal(err)
		}
	}

	t.Cleanup(func() {
		_ = ic.Close()
	})
}

// home, err := os.UserHomeDir()
// if err != nil {
// 	panic(err)
// }
// return filepath.Join(home, ".ibctest", "databases", "block.db")
