package suite

import (
	"encoding/json"
	"fmt"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	feepaytypes "github.com/CosmosContracts/juno/v30/x/feepay/types"
)

func (s *E2ETestSuite) RegisterFeePay(chain *cosmos.CosmosChain, user ibc.Wallet, contract string, walletLimit int) {
	t := s.T()
	feePayContract := &feepaytypes.FeePayContract{
		ContractAddress: contract,
		Balance:         0,
		WalletLimit:     uint64(walletLimit),
	}

	metadataJSON, err := json.Marshal(feePayContract)
	require.NoError(t, err)

	cmd := []string{
		"junod", "tx", "feepay", "register", user.FormattedAddress(), string(metadataJSON),
		"--home", chain.HomeDir(),
		"--node", chain.GetRPCAddress(),
		"--chain-id", chain.Config().ChainID,
		"--fees", "500ujuno",
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	}
	stdout, _, err := chain.Exec(s.Ctx, cmd, nil)
	require.NoError(t, err)

	s.DebugOutput(string(stdout))

	err = testutil.WaitForBlocks(s.Ctx, 2, chain)
	require.NoError(t, err)
}

func (s *E2ETestSuite) FundFeePayContract(chain *cosmos.CosmosChain, user ibc.Wallet, contract string, amountCoin string) {
	t := s.T()
	cmd := []string{
		"junod", "tx", "feepay", "fund", user.FormattedAddress(), contract, amountCoin,
		"--home", chain.HomeDir(),
		"--node", chain.GetRPCAddress(),
		"--chain-id", chain.Config().ChainID,
		"--fees", "500ujuno",
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	}
	stdout, _, err := chain.Exec(s.Ctx, cmd, nil)
	require.NoError(t, err)

	s.DebugOutput(string(stdout))

	err = testutil.WaitForBlocks(s.Ctx, 2, chain)
	require.NoError(t, err)
}

func (s *E2ETestSuite) UpdateFeePayWalletLimit(chain *cosmos.CosmosChain, user ibc.Wallet, contract string, newLimit uint64) {
	t := s.T()
	cmd := []string{
		"junod", "tx", "feepay", "update-wallet-limit", user.FormattedAddress(), contract, fmt.Sprintf("%d", newLimit),
		"--home", chain.HomeDir(),
		"--node", chain.GetRPCAddress(),
		"--chain-id", chain.Config().ChainID,
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	}
	stdout, _, err := chain.Exec(s.Ctx, cmd, nil)
	require.NoError(t, err)

	s.DebugOutput(string(stdout))

	err = testutil.WaitForBlocks(s.Ctx, 2, chain)
	require.NoError(t, err)
}
