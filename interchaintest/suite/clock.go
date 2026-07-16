package suite

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/cosmos/interchaintest/v10/testutil"
	"github.com/stretchr/testify/require"
)

// RegisterClockContract registers a clock contract
func (s *E2ETestSuite) RegisterClockContract(chain *cosmos.CosmosChain, user ibc.Wallet, contract string) (*sdk.TxResponse, error) {
	t := s.T()
	cmd := []string{
		"clock", "register", user.FormattedAddress(), contract,
		"--home", chain.HomeDir(),
		"--fees", "50000ujuno",
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	}
	node := chain.GetNode()
	txHash, err := node.ExecTx(s.Ctx, user.KeyName(), cmd...)
	if err != nil {
		t.Fatal(err)
	}
	// convert stdout into a TxResponse
	txRes, err := chain.GetTransaction(txHash)
	if err != nil {
		t.Fatal(err)
	}

	err = testutil.WaitForBlocks(s.Ctx, 2, chain)
	require.NoError(t, err)

	return txRes, nil
}

// UnregisterClockContract unregisters a clock contract
func (s *E2ETestSuite) UnregisterClockContract(chain *cosmos.CosmosChain, user ibc.Wallet, contract string) (*sdk.TxResponse, error) {
	t := s.T()
	cmd := []string{
		"clock", "unregister", user.FormattedAddress(), contract,
		"--home", chain.HomeDir(),
		"--fees", "50000ujuno",
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	}
	node := chain.GetNode()
	txHash, err := node.ExecTx(s.Ctx, user.KeyName(), cmd...)
	if err != nil {
		t.Fatal(err)
	}
	// convert stdout into a TxResponse
	txRes, err := chain.GetTransaction(txHash)
	if err != nil {
		t.Fatal(err)
	}

	err = testutil.WaitForBlocks(s.Ctx, 2, chain)
	require.NoError(t, err)

	return txRes, nil
}

// UnjailClockContract unjails a clock contract
func (s *E2ETestSuite) UnjailClockContract(chain *cosmos.CosmosChain, user ibc.Wallet, contract string) (*sdk.TxResponse, error) {
	t := s.T()
	cmd := []string{
		"clock", "unjail", user.FormattedAddress(), contract,
		"--home", chain.HomeDir(),
		"--fees", "50000ujuno",
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	}
	node := chain.GetNode()
	txHash, err := node.ExecTx(s.Ctx, user.KeyName(), cmd...)
	if err != nil {
		t.Fatal(err)
	}
	// convert stdout into a TxResponse
	txRes, err := chain.GetTransaction(txHash)
	if err != nil {
		t.Fatal(err)
	}

	err = testutil.WaitForBlocks(s.Ctx, 2, chain)
	require.NoError(t, err)

	return txRes, nil
}

// GetClockContract retrieves a clock contract
func (s *E2ETestSuite) GetClockContract(chain *cosmos.CosmosChain, contract string) ClockContract {
	t := s.T()
	var res ClockContract

	cmd := getClockQueryCommand(contract, chain)
	node := chain.GetNode()
	stdout, _, err := node.ExecQuery(s.Ctx, cmd...)
	require.NoError(t, err)

	_ = string(stdout) // Debug output removed, use debugger if needed

	if err := json.Unmarshal(stdout, &res); err != nil {
		t.Fatal(err)
	}

	return res
}

// ValidateNoClockContract validates that a contract is not registered with the clock module
func (s *E2ETestSuite) ValidateNoClockContract(chain *cosmos.CosmosChain, contract string) {
	t := s.T()
	cmd := getClockQueryCommand(contract, chain)
	_, _, err := chain.Exec(s.Ctx, cmd, nil)
	require.Error(t, err)
}

func (s *E2ETestSuite) GetClockContractValue(chain *cosmos.CosmosChain, contract string) ClockContractResponse {
	t := s.T()
	var res ClockContractResponse
	err := chain.QueryContract(s.Ctx, contract, ContractQueryMsg{GetConfig: &struct{}{}}, &res)
	require.NoError(t, err)
	return res
}

// Get the clock query command
func getClockQueryCommand(contract string, chain *cosmos.CosmosChain) []string {
	return []string{
		"clock", "contract", contract,
		"--output", "json",
		"--node", chain.GetRPCAddress(),
	}
}
