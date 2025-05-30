package suite

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	"github.com/docker/docker/client"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/relayer"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	abcitypes "github.com/cometbft/cometbft/abci/types"
)

// ChainConstructor returns a main chain, as well as any additionally specified chains
// that are needed for tests. The first chain returned will be the chain that is used as main chain in
// e2e tests.
type ChainConstructor func(t *testing.T, specs []*interchaintest.ChainSpec, gasPrices string) []*cosmos.CosmosChain

// InterchainConstructor returns an interchain that will be used in e2e tests.
// The chains used in the interchain constructor should be the chains constructed via the ChainConstructor
type InterchainConstructor func(ctx context.Context, t *testing.T, chains []*cosmos.CosmosChain) (*interchaintest.Interchain, *client.Client, ibc.Relayer)

type KeyringOverride struct {
	keyringOptions keyring.Option
	cdc            codec.Codec
}

type TestTxConfig struct {
	SmallSendsNum          int
	LargeSendsNum          int
	TargetIncreaseGasPrice math.LegacyDec
}

func (tx *TestTxConfig) Validate() error {
	if tx.SmallSendsNum < 1 || tx.LargeSendsNum < 1 {
		return fmt.Errorf("sends num should be greater than 1")
	}

	if tx.TargetIncreaseGasPrice.IsNil() {
		return fmt.Errorf("target increase gas price is nil")
	}

	if tx.TargetIncreaseGasPrice.LTE(math.LegacyZeroDec()) {
		return fmt.Errorf("target increase gas price is less than or equal to 0")
	}

	return nil
}

// DefaultChainConstructor is the default construct of a chain that will be
// used in e2e tests. There is only a single chain that is created.
func DefaultChainConstructor(t *testing.T, spec []*interchaintest.ChainSpec, _ string) []*cosmos.CosmosChain {
	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{spec[0]})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	// require that the chain is a cosmos chain
	require.Len(t, chains, 1)
	chain := chains[0]

	cosmosChain, ok := chain.(*cosmos.CosmosChain)
	require.True(t, ok)

	return []*cosmos.CosmosChain{cosmosChain}
}

func MultipleChainsConstructor(
	t *testing.T,
	specs []*interchaintest.ChainSpec,
	_ string,
) []*cosmos.CosmosChain {
	// spin up one chain per spec
	cf := interchaintest.NewBuiltinChainFactory(
		zaptest.NewLogger(t),
		specs,
	)

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)
	require.Len(t, chains, len(specs), "expected one chain per spec")

	// cast and collect
	var cosmosChains []*cosmos.CosmosChain
	for i, chain := range chains {
		cosmosChain, ok := chain.(*cosmos.CosmosChain)
		require.Truef(t, ok, "chain[%d] is not a CosmosChain", i)
		cosmosChains = append(cosmosChains, cosmosChain)
	}

	return cosmosChains
}

// DefaultInterchainConstructor is the default constructor of an interchain that will be used in e2e tests.
func DefaultInterchainConstructor(ctx context.Context, t *testing.T, chains []*cosmos.CosmosChain) (*interchaintest.Interchain, *client.Client, ibc.Relayer) {
	require.Len(t, chains, 1)

	ic := interchaintest.NewInterchain()
	ic.AddChain(chains[0])

	// create docker network
	client, networkID := interchaintest.DockerSetup(t)

	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	// build the interchain
	err := ic.Build(ctx, nil, interchaintest.InterchainBuildOptions{
		SkipPathCreation: true,
		Client:           client,
		NetworkID:        networkID,
		TestName:         t.Name(),
	})
	require.NoError(t, err)

	return ic, client, nil
}

// FourChainInterchainConstructor is the interchain constructor that spins up 4 chains, sets up relayers, and configures IBC connections between chains.
func FourChainInterchainConstructor(ctx context.Context, t *testing.T, chains []*cosmos.CosmosChain) (*interchaintest.Interchain, *client.Client, ibc.Relayer) {
	require.Len(t, chains, 4)

	const pathAB = "ab"
	const pathBC = "bc"
	const pathCD = "cd"

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	// create docker network
	client, networkId := interchaintest.DockerSetup(t)

	relayerType, relayerName := ibc.CosmosRly, "rly"
	// Get a relayer instance
	rf := interchaintest.NewBuiltinRelayerFactory(
		relayerType,
		zaptest.NewLogger(t),
		relayer.StartupFlags("--processor", "events", "--block-history", "100"),
	)
	r := rf.Build(t, client, networkId)

	ic := interchaintest.NewInterchain()
	for _, chain := range chains {
		ic.AddChain(chain)
	}
	ic.AddRelayer(r, relayerName)
	ic.AddLink(interchaintest.InterchainLink{
		Chain1:  chains[0],
		Chain2:  chains[1],
		Relayer: r,
		Path:    pathAB,
	})
	ic.AddLink(interchaintest.InterchainLink{
		Chain1:  chains[1],
		Chain2:  chains[2],
		Relayer: r,
		Path:    pathBC,
	})
	ic.AddLink(interchaintest.InterchainLink{
		Chain1:  chains[2],
		Chain2:  chains[3],
		Relayer: r,
		Path:    pathCD,
	})

	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	// build the interchain
	err := ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		SkipPathCreation:  false,
		Client:            client,
		NetworkID:         networkId,
		TestName:          t.Name(),
		BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),
	})
	require.NoError(t, err)

	return ic, client, r
}

// TwoChainInterchainConstructor is the interchain constructor that spins up 4 chains, sets up relayers, and configures IBC connections between chains.
func TwoChainInterchainConstructor(ctx context.Context, t *testing.T, chains []*cosmos.CosmosChain) (*interchaintest.Interchain, *client.Client, ibc.Relayer) {
	require.Len(t, chains, 2)

	const pathAB = "ab"

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	// create docker network
	client, networkId := interchaintest.DockerSetup(t)

	relayerType, relayerName := ibc.CosmosRly, "rly"
	// Get a relayer instance
	rf := interchaintest.NewBuiltinRelayerFactory(
		relayerType,
		zaptest.NewLogger(t),
		relayer.StartupFlags("--processor", "events", "--block-history", "100"),
	)
	r := rf.Build(t, client, networkId)

	ic := interchaintest.NewInterchain()
	for _, chain := range chains {
		ic.AddChain(chain)
	}
	ic.AddRelayer(r, relayerName)
	ic.AddLink(interchaintest.InterchainLink{
		Chain1:  chains[0],
		Chain2:  chains[1],
		Relayer: r,
		Path:    pathAB,
	})

	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	// build the interchain
	err := ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		SkipPathCreation:  false,
		Client:            client,
		NetworkID:         networkId,
		TestName:          t.Name(),
		BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),
	})
	require.NoError(t, err)

	return ic, client, r
}

// AlphaString returns a random lowercase string of length n, locking around the global RNG
func AlphaString(n int) string {
	mu.Lock()
	defer mu.Unlock()

	const letters = "abcdefghijklmnopqrstuvwxyz"
	buf := make([]byte, n)
	for i := range buf {
		buf[i] = letters[random.IntN(len(letters))]
	}
	return string(buf)
}

// AttributeValue returns an event attribute value given the eventType and attribute key tuple.
// In the event of duplicate types and keys, returns the first attribute value found.
// If not found, returns empty string and false.
func AttributeValue(events []abcitypes.Event, eventType, attrKey string) (string, bool) {
	for _, event := range events {
		if event.Type != eventType {
			continue
		}
		for _, attr := range event.Attributes {
			if attr.Key == attrKey {
				return attr.Value, true
			}

			// tendermint < v0.37-alpha returns base64 encoded strings in events.
			key, err := base64.StdEncoding.DecodeString(attr.Key)
			if err != nil {
				continue
			}
			if string(key) == attrKey {
				value, err := base64.StdEncoding.DecodeString(attr.Value)
				if err != nil {
					continue
				}
				return string(value), true
			}
		}
	}
	return "", false
}

type safePCG struct {
	mu  sync.Mutex
	pcg *rand.PCG
}

func (s *safePCG) Uint64() uint64 {
	s.mu.Lock()
	v := s.pcg.Uint64()
	s.mu.Unlock()
	return v
}
