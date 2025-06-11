package suite

import (
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

func (s *E2ETestSuite) RegisterFeeShare(chain *cosmos.CosmosChain, user ibc.Wallet, contract, withdrawAddr string) {
	t := s.T()
	// TF gas to create cost 2mil, so we set to 2.5 to be safe
	cmd := []string{
		"junod", "tx", "feeshare", "register", contract, user.FormattedAddress(), withdrawAddr,
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
