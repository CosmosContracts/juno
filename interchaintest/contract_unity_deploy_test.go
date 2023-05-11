package interchaintest

import (
	"fmt"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"

	helpers "github.com/CosmosContracts/juno/tests/interchaintest/helpers"
)

// TestJunoUnityContractDeploy test to ensure the contract withdraw function works as expected on chain.
// - https://github.com/CosmosContracts/cw-unity-prop
func TestJunoUnityContractDeploy(t *testing.T) {
	t.Parallel()

	// Base setup
	chains := CreateThisBranchChain(t, 1, 0)
	ic, ctx, _, _ := BuildInitialChain(t, chains)

	// Chains
	juno := chains[0].(*cosmos.CosmosChain)
	nativeDenom := juno.Config().Denom

	// Users
	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", int64(10_000_000), juno, juno)
	user := users[0]
	withdrawUser := users[1]
	withdrawAddr := withdrawUser.FormattedAddress()

	// TEST DEPLOY (./scripts/deploy_ci.sh)
	// Upload & init unity contract with no admin in test mode
	msg := fmt.Sprintf(`{"native_denom":"%s","withdraw_address":"%s","withdraw_delay_in_days":28}`, nativeDenom, withdrawAddr)
	_, contractAddr := helpers.SetupContract(t, ctx, juno, user.KeyName(), "contracts/cw_unity_prop.wasm", msg)
	t.Log("testing Unity contractAddr", contractAddr)

	// Execute to start the withdrawl countdown
	juno.ExecuteContract(ctx, withdrawUser.KeyName(), contractAddr, `{"start_withdraw":{}}`)

	// make a query with GetUnityContractWithdrawalReadyTime
	res := helpers.GetUnityContractWithdrawalReadyTime(t, ctx, juno, contractAddr)
	t.Log("WithdrawalReadyTimestamp", res.Data.WithdrawalReadyTimestamp)

	t.Cleanup(func() {
		_ = ic.Close()
	})
}
