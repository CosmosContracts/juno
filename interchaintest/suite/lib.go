package suite

import (
	"context"
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/docker/docker/client"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// ChainConstructor returns a main chain, as well as any additionally specified chains
// that are needed for tests. The first chain returned will be the chain that is used as main chain in
// e2e tests.
type ChainConstructor func(t *testing.T, spec *interchaintest.ChainSpec, gasPrices string) []*cosmos.CosmosChain

// InterchainConstructor returns an interchain that will be used in e2e tests.
// The chains used in the interchain constructor should be the chains constructed via the ChainConstructor
type InterchainConstructor func(ctx context.Context, t *testing.T, chains []*cosmos.CosmosChain) (*interchaintest.Interchain, *client.Client)

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
func DefaultChainConstructor(t *testing.T, spec *interchaintest.ChainSpec, _ string) []*cosmos.CosmosChain {
	// require that NumFullNodes == NumValidators == 4
	require.Equal(t, 4, *spec.NumValidators)

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{spec})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	// require that the chain is a cosmos chain
	require.Len(t, chains, 1)
	chain := chains[0]

	cosmosChain, ok := chain.(*cosmos.CosmosChain)
	require.True(t, ok)

	return []*cosmos.CosmosChain{cosmosChain}
}

// DefaultInterchainConstructor is the default constructor of an interchain that will be used in e2e tests.
func DefaultInterchainConstructor(ctx context.Context, t *testing.T, chains []*cosmos.CosmosChain) (*interchaintest.Interchain, *client.Client) {
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

	return ic, client
}
