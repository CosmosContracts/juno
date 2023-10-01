package interchaintest

import (
	"fmt"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"

	helpers "github.com/CosmosContracts/juno/tests/interchaintest/helpers"
)

// TestJunoCwHooks
func TestJunoCwHooks(t *testing.T) {
	t.Parallel()

	cfg := junoConfig

	// Base setup
	chains := CreateChainWithCustomConfig(t, 1, 0, cfg)
	ic, ctx, _, _ := BuildInitialChain(t, chains)

	// Chains
	juno := chains[0].(*cosmos.CosmosChain)

	// Users
	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", int64(10_000_000_000), juno, juno)
	user := users[0]

	// Upload & init contract payment to another address
	// TODO: convert this to a contract with staking sudo msgs
	_, contractAddr := helpers.SetupContract(t, ctx, juno, user.KeyName(), "contracts/clock_example.wasm", `{}`)

	// register contract addr with the command

	helpers.RegisterCwHooksStaking(t, ctx, juno, user, contractAddr)

	c := helpers.GetCwHooksStakingContracts(t, ctx, juno)
	fmt.Printf("c: %v\n", c)

	// do a staking action here, and confirm it works and is modified in the contract.
	// TODO: do for all actions, staking and gov.

	t.Cleanup(func() {
		_ = ic.Close()
	})
}
