package fees_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8"
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

	// Upload & init contract payment to another address
	codeId, err := s.Chain.StoreContract(s.Ctx, admin.KeyName(), "../../contracts/cw_template.wasm", "--fees", "50000ujuno")
	if err != nil {
		t.Fatal(err)
	}

	contractAddr, err := s.Chain.InstantiateContract(s.Ctx, admin.KeyName(), codeId, `{"count":0}`, true)
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
	require.Equal(t, beforeContract.FeePayContract.Balance, strconv.Itoa(balance))
	require.Equal(t, beforeContract.FeePayContract.WalletLimit, strconv.Itoa(int(limit)))

	// execute it from another account with enough fees (standard Tx)
	txHash, err := s.Chain.ExecuteContract(s.Ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--fees", "500"+nativeDenom)
	require.NoError(err)
	fmt.Println("txHash", txHash)

	beforeBal, err := s.Chain.GetBalance(s.Ctx, user.FormattedAddress(), nativeDenom)
	require.NoError(err)

	// execute it from another account and have the dev pay it
	txHash, err = s.Chain.ExecuteContract(s.Ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--fees", "0"+nativeDenom)
	require.NoError(err)
	fmt.Println("txHash", txHash)

	afterBal, err := s.Chain.GetBalance(s.Ctx, user.FormattedAddress(), nativeDenom)
	require.NoError(err)

	// validate users balance did not change
	require.Equal(t, beforeBal, afterBal)

	// validate the contract balance went down
	afterContract, err := s.QueryClients.FeepayClient.FeePayContract(
		s.Ctx,
		&types.QueryFeePayContractRequest{
			ContractAddress: contractAddr,
		},
	)
	t.Log("afterContract", afterContract)
	require.Equal(t, afterContract.FeePayContract.Balance, strconv.Itoa(balance-500))

	uses, err := s.QueryClients.FeepayClient.FeePayContractUses(
		s.Ctx,
		&types.QueryFeePayContractUsesRequest{
			ContractAddress: contractAddr,
			WalletAddress:   user.FormattedAddress(),
		},
	)
	t.Log("uses", uses)
	require.Equal(t, uses.Uses, "1")

	// Instantiate a new contract
	contractAddr, err = s.Chain.InstantiateContract(s.Ctx, admin.KeyName(), codeId, `{"count":0}`, true)
	if err != nil {
		t.Fatal(err)
	}

	// Succeed - Test a regular CW contract with fees, regular sdk logic handles Tx
	txHash, err = s.Chain.ExecuteContract(s.Ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--fees", "500"+nativeDenom)
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
	txHash, err = s.Chain.ExecuteContract(s.Ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--fees", "500"+nativeDenom)
	require.NoError(err)
	fmt.Println("txHash", txHash)

	// Before balance - should be the same as after balance (feepay covers fee)
	// Calculated before interacting with a registered contract to ensure the
	// contract covers the fee.
	beforeBal, err = s.Chain.GetBalance(s.Ctx, user.FormattedAddress(), nativeDenom)
	require.NoError(err)

	// Test the registered FeePay contract - without providing fees
	txHash, err = s.Chain.ExecuteContract(s.Ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--fees", "0"+nativeDenom)
	require.NoError(err)
	fmt.Println("txHash", txHash)

	// After balance
	afterBal, err = s.Chain.GetBalance(s.Ctx, user.FormattedAddress(), nativeDenom)
	require.NoError(err)

	// Validate users balance did not change
	require.Equal(t, beforeBal, afterBal)

	// Test the fallback sdk route is triggered when the FeePay Tx fails
	// Fail - Test the registered contract - without fees, exceeded wallet limit
	txHash, err = s.Chain.ExecuteContract(s.Ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--fees", "0"+nativeDenom)
	require.Error(err)
	fmt.Println("txHash", txHash)

	// Test the registered contract - without fees, but specified gas
	// Tx should succeed, because it uses the sdk fallback route
	txHash, err = s.Chain.ExecuteContract(s.Ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--gas", "200000")
	require.NoError(err)
	fmt.Println("txHash", txHash)
}
