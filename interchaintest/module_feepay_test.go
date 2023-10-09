package interchaintest

import (
	"fmt"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"

	helpers "github.com/CosmosContracts/juno/tests/interchaintest/helpers"
)

// TestJunoFeePay
func TestJunoFeePay(t *testing.T) {
	t.Parallel()

	cfg := junoConfig
	cfg.GasPrices = "0.0025ujuno"

	// Base setup
	chains := CreateChainWithCustomConfig(t, 1, 0, cfg)
	ic, ctx, _, _ := BuildInitialChain(t, chains)

	// Chains
	juno := chains[0].(*cosmos.CosmosChain)

	nativeDenom := juno.Config().Denom

	// Users
	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", int64(10_000_000), juno, juno)
	admin := users[0]
	user := users[1]

	// Upload & init contract payment to another address
	codeId, err := juno.StoreContract(ctx, admin.KeyName(), "contracts/cw_template.wasm", "--fees", "50000ujuno")
	if err != nil {
		t.Fatal(err)
	}

	contractAddr, err := juno.InstantiateContract(ctx, admin.KeyName(), codeId, `{"count":0}`, true)
	if err != nil {
		t.Fatal(err)
	}

	// Register contract for 0 fee usage (x amount of times)
	helpers.RegisterFeePay(t, ctx, juno, admin, contractAddr, 5)
	helpers.FundFeePayContract(t, ctx, juno, admin, contractAddr, "1000000"+nativeDenom)

	// execute against it from another account with enough fees (standard Tx)
	txHash, err := juno.ExecuteContract(ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--fees", "500"+nativeDenom)
	if err != nil {
		// TODO:
		t.Log(err)
	}
	fmt.Println("txHash", txHash)

	// execute against it from another account and have the dev pay it
	txHash, err = juno.ExecuteContract(ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--fees", "0"+nativeDenom)
	if err != nil {
		// TODO:
		t.Log(err)
	}
	fmt.Println("txHash", txHash)

	// validate their balance did not go down, and that the contract did infact increase +=1
	// if balance, err := juno.GetBalance(ctx, feeRcvAddr, nativeDenom); err != nil {
	// 	t.Fatal(err)
	// } else if balance != 0 {
	// 	t.Fatal("balance not 0")
	// }

	// wait blocks
	testutil.WaitForBlocks(ctx, 200, juno)

	t.Cleanup(func() {
		_ = ic.Close()
	})
}
