package interchaintest

import (
	"fmt"
	"strconv"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/stretchr/testify/assert"

	helpers "github.com/CosmosContracts/juno/tests/interchaintest/helpers"
)

// TestJunoBurnModule ensures the junoburn module register and execute sharing functions work properly on smart contracts.
// This is required due to how x/mint handles minting tokens for the target supply.
// It is purely for developers ::BurnTokens to function as expected.
func TestJunoBurnModule(t *testing.T) {
	t.Parallel()

	// Base setup
	chains := CreateThisBranchChain(t, 1, 0)
	ic, ctx, _, _ := BuildInitialChain(t, chains)

	// Chains
	juno := chains[0].(*cosmos.CosmosChain)

	nativeDenom := juno.Config().Denom

	// Users
	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", sdkmath.NewInt(10_000_000), juno, juno)
	user := users[0]

	// Upload & init contract
	_, contractAddr := helpers.SetupContract(t, ctx, juno, user.KeyName(), "contracts/cw_testburn.wasm", `{}`)

	// get balance before execute
	balance, err := juno.GetBalance(ctx, user.FormattedAddress(), nativeDenom)
	if err != nil {
		t.Fatal(err)
	}

	// execute burn of tokens
	burnAmt := int64(1_000_000)
	helpers.ExecuteMsgWithAmount(t, ctx, juno, user, contractAddr, strconv.Itoa(int(burnAmt))+nativeDenom, `{"burn_token":{}}`)

	// verify it is down 1_000_000 tokens since the burn
	updatedBal, err := juno.GetBalance(ctx, user.FormattedAddress(), nativeDenom)
	if err != nil {
		t.Fatal(err)
	}

	// Verify the funds were sent, and burned.
	fmt.Println(balance, updatedBal)
	assert.Equal(t, burnAmt, balance.Sub(updatedBal).Int64(), fmt.Sprintf("balance should be %d less than updated balance", burnAmt))

	t.Cleanup(func() {
		_ = ic.Close()
	})
}
