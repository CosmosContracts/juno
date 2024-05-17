package interchaintest

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	sdkmath "cosmossdk.io/math"
	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"

	helpers "github.com/CosmosContracts/juno/tests/interchaintest/helpers"
)

// TestJunoDrip ensures the drip module properly distributes tokens from whitelisted accounts.
func TestJunoDrip(t *testing.T) {
	t.Parallel()

	// Setup new pre determined user (from test_node.sh)
	mnemonic := "decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry"
	addr := "juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl"

	// Base setup
	newCfg := junoConfig
	newCfg.ModifyGenesis = cosmos.ModifyGenesis(append(defaultGenesisKV, []cosmos.GenesisKV{
		{
			Key:   "app_state.drip.params.allowed_addresses",
			Value: []string{addr},
		},
	}...))

	chains := CreateChainWithCustomConfig(t, 1, 0, newCfg)
	ic, ctx, _, _ := BuildInitialChain(t, chains)

	// Chains
	juno := chains[0].(*cosmos.CosmosChain)
	nativeDenom := juno.Config().Denom

	// User
	user, err := interchaintest.GetAndFundTestUserWithMnemonic(ctx, "default", mnemonic, sdkmath.NewInt(1_000_000_000_000), juno)
	if err != nil {
		t.Fatal(err)
	}

	// New TF token to distributes
	tfDenom := helpers.CreateTokenFactoryDenom(t, ctx, juno, user, "dripme", fmt.Sprintf("0%s", Denom))
	distributeAmt := math.NewInt(1_000_000)
	helpers.MintTokenFactoryDenom(t, ctx, juno, user, distributeAmt.Uint64(), tfDenom)

	// Stake some tokens
	vals := helpers.GetValidators(t, ctx, juno)
	valoper := vals.Validators[0].OperatorAddress

	stakeAmt := sdkmath.NewInt(100_000_000_000)
	helpers.StakeTokens(t, ctx, juno, user, valoper, fmt.Sprintf("%d%s", stakeAmt, nativeDenom))

	// Drip the TF Tokens to all stakers
	distribute := sdkmath.NewInt(1_000_000)
	helpers.DripTokens(t, ctx, juno, user, fmt.Sprintf("%d%s", distribute, tfDenom))

	// Claim staking rewards to capture the drip
	helpers.ClaimStakingRewards(t, ctx, juno, user, valoper)

	// Check balances has the TF Denom from the claim
	bals, _ := juno.AllBalances(ctx, user.FormattedAddress())
	fmt.Println("balances", bals)

	found := false
	for _, bal := range bals {
		if bal.Denom == tfDenom {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("did not find drip token")
	}

	t.Cleanup(func() {
		_ = ic.Close()
	})
}
