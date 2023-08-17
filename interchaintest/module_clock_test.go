package interchaintest

import (
	"testing"

	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"

	helpers "github.com/CosmosContracts/juno/tests/interchaintest/helpers"
)

// TestJunoClock ensures the clock module auto executes allowed contracts.
func TestJunoClock(t *testing.T) {
	t.Parallel()

	cfg := junoConfig
	// set allowed address by default to the contract creator
	// cfg.ModifyGenesis = cosmos.ModifyGenesis(append(defaultGenesisKV, []cosmos.GenesisKV{

	// Base setup
	chains := CreateChainWithCustomConfig(t, 1, 0, cfg)
	ic, ctx, _, _ := BuildInitialChain(t, chains)

	// Chains
	juno := chains[0].(*cosmos.CosmosChain)

	// Users
	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", int64(10_000_000), juno, juno)
	user := users[0]

	// Upload & init contract payment to another address
	_, _ = helpers.SetupContract(t, ctx, juno, user.KeyName(), "contracts/clock_example.wasm", `{}`)

	// wait 1 block
	// query the contractAddress config & see if it increased.
	// pub struct Config {
	// 	pub val: u32,
	// }

	// TODO: param proposal to add the contract address to the allowed list? or remove it and see if it still increments.

	t.Cleanup(func() {
		_ = ic.Close()
	})
}
