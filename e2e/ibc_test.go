package e2e

import (
	"context"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v3"
	"github.com/strangelove-ventures/interchaintest/v3/ibc"
	"github.com/strangelove-ventures/interchaintest/v3/testreporter"
	"go.uber.org/zap/zaptest"
)

// go test -timeout 99999s -run ^TestLearn$ github.com/CosmosContracts/juno/v13/ibctest -v
func TestRun(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// How to use this branches latest commit for builds? where we build the docker first then use it here?
	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:      "juno",
			ChainName: "juno1",
			Version:   "latest",
			ChainConfig: ibc.ChainConfig{
				GasPrices:     "0ujuno",
				GasAdjustment: 2.0,
			},
		},
		{
			Name:      "juno",
			ChainName: "juno2",
			Version:   "latest",
			ChainConfig: ibc.ChainConfig{
				GasPrices:     "0ujuno",
				GasAdjustment: 2.0,
			},
		},
	})

	chains, err := cf.Chains(t.Name())
	if err != nil {
		t.Fatal(err)
	}

	left, right := chains[0], chains[1]

	client, network := interchaintest.DockerSetup(t)
	relayer := interchaintest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		zaptest.NewLogger(t),
	).Build(t, client, network)

	const ibcPath = "juno-juno"
	ic := interchaintest.NewInterchain().
		AddChain(left).
		AddChain(right).
		AddRelayer(relayer, "relayer").
		AddLink(interchaintest.InterchainLink{
			Chain1:  left,
			Chain2:  right,
			Relayer: relayer,
			Path:    ibcPath,
		})

	// NopReporter doesn't write to a log file.
	erp := testreporter.NewNopReporter().RelayerExecReporter(t)
	err = ic.Build(ctx, erp, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = ic.Close()
	})

	err = relayer.StartRelayer(ctx, erp, ibcPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		err := relayer.StopRelayer(ctx, erp)
		if err != nil {
			t.Logf("couldn't stop relayer: %s", err)
		}
	})

	// users := interchaintest.GetAndFundTestUsers(t, ctx, "default", int64(10_000_000), left, right)
	// leftUser := users[0]
	// rightUser := users[1]

	// leftCosmosChain := left.(*cosmos.CosmosChain)
	// rightCosmosChain := right.(*cosmos.CosmosChain)

	// TODO: IBC Transfer from chainA to B
}
