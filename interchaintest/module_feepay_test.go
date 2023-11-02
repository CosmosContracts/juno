package interchaintest

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	helpers "github.com/CosmosContracts/juno/tests/interchaintest/helpers"
)

// TestJunoFeePay
func TestJunoFeePay(t *testing.T) {
	t.Parallel()

	cfg := junoConfig
	cfg.GasPrices = "0.0025ujuno"

	// 0.002500000000000000
	coin := sdk.NewDecCoinFromDec(cfg.Denom, sdk.NewDecWithPrec(25, 4))
	cfg.ModifyGenesis = cosmos.ModifyGenesis(append(defaultGenesisKV, []cosmos.GenesisKV{
		{
			Key:   "app_state.globalfee.params.minimum_gas_prices",
			Value: sdk.DecCoins{coin},
		},
		{
			// override default impl.
			Key:   "app_state.feepay.params.enable_feepay",
			Value: true,
		},
	}...))

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
	limit := 5
	balance := 1_000_000
	helpers.RegisterFeePay(t, ctx, juno, admin, contractAddr, limit)
	helpers.FundFeePayContract(t, ctx, juno, admin, contractAddr, strconv.Itoa(balance)+nativeDenom)

	beforeContract := helpers.GetFeePayContract(t, ctx, juno, contractAddr)
	t.Log("beforeContract", beforeContract)
	require.Equal(t, beforeContract.FeePayContract.Balance, strconv.Itoa(balance))
	require.Equal(t, beforeContract.FeePayContract.WalletLimit, strconv.Itoa(int(limit)))

	// execute against it from another account with enough fees (standard Tx)
	txHash, err := juno.ExecuteContract(ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--fees", "500"+nativeDenom)
	require.NoError(t, err)
	fmt.Println("txHash", txHash)

	beforeBal, err := juno.GetBalance(ctx, user.FormattedAddress(), nativeDenom)
	require.NoError(t, err)

	// execute against it from another account and have the dev pay it
	txHash, err = juno.ExecuteContract(ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--fees", "0"+nativeDenom)
	require.NoError(t, err)
	fmt.Println("txHash", txHash)

	afterBal, err := juno.GetBalance(ctx, user.FormattedAddress(), nativeDenom)
	require.NoError(t, err)

	// validate users balance did not change
	require.Equal(t, beforeBal, afterBal)

	// validate the contract balance went down
	afterContract := helpers.GetFeePayContract(t, ctx, juno, contractAddr)
	t.Log("afterContract", afterContract)
	require.Equal(t, afterContract.FeePayContract.Balance, strconv.Itoa(balance-500))

	uses := helpers.GetFeePayUses(t, ctx, juno, contractAddr, user.FormattedAddress())
	t.Log("uses", uses)
	require.Equal(t, uses.Uses, "1")

	// Instantiate a new contract
	contractAddr, err = juno.InstantiateContract(ctx, admin.KeyName(), codeId, `{"count":0}`, true)
	if err != nil {
		t.Fatal(err)
	}

	// Test a contract that is uploaded and not registered - with fees
	txHash, err = juno.ExecuteContract(ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--fees", "500"+nativeDenom)
	require.NoError(t, err)
	fmt.Println("txHash", txHash)

	// Fail - Test a contract that is uploaded and not registered - without fees
	txHash, err = juno.ExecuteContract(ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--fees", "0"+nativeDenom)
	require.Error(t, err)
	fmt.Println("txHash", txHash)

	// Register the new contract with a limit of 1, fund contract
	helpers.RegisterFeePay(t, ctx, juno, admin, contractAddr, 1)
	helpers.FundFeePayContract(t, ctx, juno, admin, contractAddr, strconv.Itoa(balance)+nativeDenom)

	// Test the registered contract - with fees
	txHash, err = juno.ExecuteContract(ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--fees", "500"+nativeDenom)
	require.NoError(t, err)
	fmt.Println("txHash", txHash)

	// Before balance - should be the same as after balance (feepay covers fee)
	beforeBal, err = juno.GetBalance(ctx, user.FormattedAddress(), nativeDenom)
	require.NoError(t, err)

	// Test the registered contract - without providing fees
	txHash, err = juno.ExecuteContract(ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--fees", "0"+nativeDenom)
	require.NoError(t, err)
	fmt.Println("txHash", txHash)

	// After balance
	afterBal, err = juno.GetBalance(ctx, user.FormattedAddress(), nativeDenom)
	require.NoError(t, err)

	// Validate users balance did not change
	require.Equal(t, beforeBal, afterBal)

	// Fail - Test the registered contract - without fees, exceeded wallet limit
	// Test the fallback sdk route is triggered when the FeePay Tx fails
	txHash, err = juno.ExecuteContract(ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--fees", "0"+nativeDenom)
	require.Error(t, err)
	fmt.Println("txHash", txHash)

	// Test the registered contract - without fees, but specified gas
	txHash, err = juno.ExecuteContract(ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--gas", "200000")
	require.NoError(t, err)
	fmt.Println("txHash", txHash)

	t.Cleanup(func() {
		_ = ic.Close()
	})
}
