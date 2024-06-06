package interchaintest

import (
	"testing"

	sdkmath "cosmossdk.io/math"

	junoconformance "github.com/CosmosContracts/juno/tests/interchaintest/conformance"
	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/stretchr/testify/require"
)

// TestBasicJunoStart is a basic test to assert that spinning up a Juno network with one validator works properly.
func TestBasicJunoStart(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	// Base setup
	chains := CreateThisBranchChain(t, 1, 0)
	ic, ctx, _, _ := BuildInitialChain(t, chains)

	chain := chains[0].(*cosmos.CosmosChain)

	userFunds := sdkmath.NewInt(10_000_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, chain)
	chainUser := users[0]

	junoconformance.ConformanceCosmWasm(t, ctx, chain, chainUser)

	require.NotNil(t, ic)
	require.NotNil(t, ctx)

	t.Cleanup(func() {
		_ = ic.Close()
	})
}
