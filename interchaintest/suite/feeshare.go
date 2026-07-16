package suite

import (
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/cosmos/interchaintest/v10/testutil"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

func (s *E2ETestSuite) RegisterFeeShare(chain *cosmos.CosmosChain, user ibc.Wallet, contract, withdrawAddr string) {
	t := s.T()
	// v30 feemarket rejects zero-fee txs ("no fee coin provided. Must provide one."),
	// so register-feeshare needs an explicit --fees just like every other ante-checked tx.
	cmd := []string{
		"junod", "tx", "feeshare", "register", contract, user.FormattedAddress(), withdrawAddr,
		"--home", chain.HomeDir(),
		"--node", chain.GetRPCAddress(),
		"--chain-id", chain.Config().ChainID,
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"--from", user.KeyName(),
		"--gas", "auto",
		"--gas-adjustment", "3",
		"--fees", "50000ujuno",
		"-y",
	}
	stdout, _, err := chain.Exec(s.Ctx, cmd, nil)
	require.NoError(t, err)

	s.DebugOutput(string(stdout))

	err = testutil.WaitForBlocks(s.Ctx, 2, chain)
	require.NoError(t, err)
}
