package interchaintest

import (
	"testing"

	"github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/testutil"
	"github.com/stretchr/testify/require"

	helpers "github.com/CosmosContracts/juno/tests/interchaintest/helpers"
)

// TestJunoTokenFactory ensures the tokenfactory module & bindings work properly
func TestJunoTokenFactory(t *testing.T) {
	t.Parallel()

	chains := CreateThisBranchChain(t)
	juno := chains[0].(*cosmos.CosmosChain)

	ic, ctx, _, _ := BuildInitialChain(t, chains)

	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", int64(10_000_000), juno, juno)
	user := users[0]
	// uaddr := user.Bech32Address(juno.Config().Bech32Prefix)

	user2 := users[1]
	// uaddr2 := user2.Bech32Address(CHAIN_PREFIX)

	tfDenom := helpers.CreateTokenFactoryDenom(t, ctx, juno, user, "ictestdenom")
	t.Log("tfDenom", tfDenom)

	helpers.MintToTokenFactoryDenom(t, ctx, juno, user, user2, 100, tfDenom)
	err := testutil.WaitForBlocks(ctx, 2, juno)
	require.NoError(t, err)
	t.Log("minted tfDenom to user")

	// ensure user2 has 100 tfDenom
	// balance, _ := juno.GetBalance(ctx, uaddr, "ujuno")
	// t.Log("balance", balance)

	// I don't think GRPC qwuery likes /'s in denoms
	// balance, _ = juno.GetBalance(ctx, uaddr, tfDenom)
	// t.Log("balance", balance)

	// upload a TF contract here & interact with it
	// Have a query which shows the contract balances held by the contract. This way we can insure it mints to itself.
	// Use same contract for a feeshare test with changing params

	t.Cleanup(func() {
		_ = ic.Close()
	})
}
