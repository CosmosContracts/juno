package fees_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/cosmos/interchaintest/v10"
	"github.com/stretchr/testify/suite"

	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"
	"github.com/CosmosContracts/juno/v30/x/feepay/types"
)

type FeesTestSuite struct {
	*e2esuite.E2ETestSuite
}

func TestFeesTestSuite(t *testing.T) {
	numValidators := 1
	numFullNodes := 0

	spec := &interchaintest.ChainSpec{
		ChainName:     "juno-fees",
		Name:          "juno",
		NumValidators: &numValidators,
		NumFullNodes:  &numFullNodes,
		Version:       e2esuite.DefaultSpec.Version,
		NoHostMount:   &e2esuite.DefaultNoHostMount,
		ChainConfig:   e2esuite.DefaultConfig,
	}

	s := e2esuite.NewE2ETestSuite(
		[]*interchaintest.ChainSpec{spec},
		e2esuite.DefaultTxCfg,
	)

	t.Parallel()
	t.Cleanup(func() {
		_ = s.Ic.Close()
	})

	testSuite := &FeesTestSuite{E2ETestSuite: s}
	suite.Run(t, testSuite)
}

// TestFeePay ensures that the x/feepay module handles external fee paying
func (s *FeesTestSuite) TestFeePay() {
	t := s.T()
	require := s.Require()
	nativeDenom := s.Chain.Config().Denom

	// Users
	user := s.GetAndFundTestUser("default", int64(10_000_000), s.Chain)
	admin := s.GetAndFundTestUser("admin", int64(10_000_000), s.Chain)

	// Upload & init contract payment to another address.
	// Wasm-store at v30 minBaseGasPrice 0.075 needs ≥225k fee for ~3M gas-limit; 1M leaves headroom.
	codeId, err := s.Chain.StoreContract(s.Ctx, admin.KeyName(), "../../contracts/cw_template.wasm", "--fees", "1000000ujuno")
	if err != nil {
		t.Fatal(err)
	}

	contractAddr, err := s.Chain.InstantiateContract(s.Ctx, admin.KeyName(), codeId, `{"count":0}`, true, "--gas", "auto", "--fees", "200000"+nativeDenom)
	if err != nil {
		t.Fatal(err)
	}

	// Register contract for 0 fee usage (x amount of times)
	limit := 5
	balance := 1_000_000
	s.RegisterFeePay(s.Chain, admin, contractAddr, limit)
	s.FundFeePayContract(s.Chain, admin, contractAddr, strconv.Itoa(balance)+nativeDenom)

	beforeContract, err := s.QueryClients.FeepayClient.FeePayContract(
		s.Ctx,
		&types.QueryFeePayContractRequest{
			ContractAddress: contractAddr,
		},
	)

	require.NoError(err)
	t.Log("beforeContract", beforeContract)
	// FeePayContract.Balance and .WalletLimit are uint64 (see x/feepay/types/feepay.pb.go);
	// the prior require.Equal calls (a) wrongly passed `t` as the first arg to a method
	// on the Assertions object and (b) compared uint64 to strconv.Itoa output. Compare
	// uint64-to-uint64.
	require.Equal(uint64(balance), beforeContract.FeePayContract.Balance)
	require.Equal(uint64(limit), beforeContract.FeePayContract.WalletLimit)

	// execute it from another account with enough fees (standard Tx)
	txHash, err := s.Chain.ExecuteContract(s.Ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--fees", "50000"+nativeDenom)
	require.NoError(err)
	fmt.Println("txHash", txHash)

	beforeBal, err := s.Chain.GetBalance(s.Ctx, user.FormattedAddress(), nativeDenom)
	require.NoError(err)

	// execute it from another account and have the dev pay it.
	// Explicit --gas: SDK default 200000 is too tight under v30 (feepay accounting
	// + wasmvm v3 increment runs ~202k); zero-fee tx can't use --gas auto reliably.
	txHash, err = s.Chain.ExecuteContract(s.Ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--gas", "500000", "--fees", "0"+nativeDenom)
	require.NoError(err)
	fmt.Println("txHash", txHash)

	afterBal, err := s.Chain.GetBalance(s.Ctx, user.FormattedAddress(), nativeDenom)
	require.NoError(err)

	// validate users balance did not change
	require.Equal(beforeBal, afterBal)

	// validate the contract balance went down — exact deduction = tx gas-limit × current gas price,
	// which is non-deterministic test-side; assert it dropped (and stays under the funded balance).
	afterContract, err := s.QueryClients.FeepayClient.FeePayContract(
		s.Ctx,
		&types.QueryFeePayContractRequest{
			ContractAddress: contractAddr,
		},
	)
	require.NoError(err)
	t.Log("afterContract", afterContract)
	require.Less(afterContract.FeePayContract.Balance, beforeContract.FeePayContract.Balance,
		"feepay should have deducted from contract balance")

	uses, err := s.QueryClients.FeepayClient.FeePayContractUses(
		s.Ctx,
		&types.QueryFeePayContractUsesRequest{
			ContractAddress: contractAddr,
			WalletAddress:   user.FormattedAddress(),
		},
	)
	require.NoError(err)
	t.Log("uses", uses)
	// Uses is uint64 (see x/feepay/types/feepay.pb.go); compare uint64-to-uint64.
	require.Equal(uint64(1), uses.Uses)

	// Instantiate a new contract
	contractAddr, err = s.Chain.InstantiateContract(s.Ctx, admin.KeyName(), codeId, `{"count":0}`, true, "--gas", "auto", "--fees", "200000"+nativeDenom)
	if err != nil {
		t.Fatal(err)
	}

	// Succeed - Test a regular CW contract with fees, regular sdk logic handles Tx
	txHash, err = s.Chain.ExecuteContract(s.Ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--fees", "50000"+nativeDenom)
	require.NoError(err)
	fmt.Println("txHash", txHash)

	// Fail - Testing an unregistered contract with no fees, FeePay Tx logic will fail it due to not being registered
	txHash, err = s.Chain.ExecuteContract(s.Ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--fees", "0"+nativeDenom)
	require.Error(err)
	fmt.Println("txHash", txHash)

	// Register the new contract with a limit of 1, fund contract
	s.RegisterFeePay(s.Chain, admin, contractAddr, 1)
	s.FundFeePayContract(s.Chain, admin, contractAddr, strconv.Itoa(balance)+nativeDenom)

	// Test the registered contract - with fees
	// Will succeed, routed through normal sdk because a fee was provided
	txHash, err = s.Chain.ExecuteContract(s.Ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--fees", "50000"+nativeDenom)
	require.NoError(err)
	fmt.Println("txHash", txHash)

	// Before balance - should be the same as after balance (feepay covers fee)
	// Calculated before interacting with a registered contract to ensure the
	// contract covers the fee.
	beforeBal, err = s.Chain.GetBalance(s.Ctx, user.FormattedAddress(), nativeDenom)
	require.NoError(err)

	// Test the registered FeePay contract - without providing fees
	txHash, err = s.Chain.ExecuteContract(s.Ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--gas", "500000", "--fees", "0"+nativeDenom)
	require.NoError(err)
	fmt.Println("txHash", txHash)

	// After balance
	afterBal, err = s.Chain.GetBalance(s.Ctx, user.FormattedAddress(), nativeDenom)
	require.NoError(err)

	// Validate users balance did not change
	require.Equal(beforeBal, afterBal)

	// Test the fallback sdk route is triggered when the FeePay Tx fails
	// Fail - Test the registered contract - without fees, exceeded wallet limit
	txHash, err = s.Chain.ExecuteContract(s.Ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--gas", "500000", "--fees", "0"+nativeDenom)
	require.Error(err)
	fmt.Println("txHash", txHash)

	// Test the registered contract - without fees, but specified gas
	// Tx should succeed, because it uses the sdk fallback route
	txHash, err = s.Chain.ExecuteContract(s.Ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--gas", "200000")
	require.NoError(err)
	fmt.Println("txHash", txHash)
}
