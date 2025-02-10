package helpers

import (
	"context"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
)

func ExecuteAuthzGrantMsg(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, granter ibc.Wallet, grantee ibc.Wallet, msgType string) {
	if !strings.HasPrefix(msgType, "/") {
		msgType = "/" + msgType
	}

	cmd := []string{
		"junod", "tx", "authz", "grant", grantee.FormattedAddress(), "generic",
		"--msg-type", msgType,
		"--node", chain.GetRPCAddress(),
		"--from", granter.KeyName(),
		"--chain-id", chain.Config().ChainID,
		"--home", chain.HomeDir(),
		"--gas", "500000",
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	}

	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	debugOutput(t, string(stdout))

	if err := testutil.WaitForBlocks(ctx, 2, chain); err != nil {
		t.Fatal(err)
	}
}

func ExecuteAuthzExecMsgWithFee(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, grantee ibc.Wallet, contractAddr, amount, feeCoin, message string) {
	// Get the node to execute the command & write output to file
	node := chain.Nodes()[0]
	filePath := "authz.json"
	generateMsg := []string{
		"junod", "tx", "wasm", "execute", contractAddr, message,
		"--home", chain.HomeDir(),
		"--chain-id", chain.Config().ChainID,
		"--node", chain.GetRPCAddress(),
		"--from", grantee.KeyName(),
		"--gas", "500000",
		"--fees", feeCoin,
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"--generate-only",
	}

	// Generate msg output
	res, resErr, err := node.Exec(ctx, generateMsg, nil)
	if resErr != nil {
		t.Fatal(resErr)
	}
	if err != nil {
		t.Fatal(err)
	}

	// Write output to file
	err = node.WriteFile(ctx, res, filePath)
	if err != nil {
		t.Fatal(err)
	}

	// Execute the command
	cmd := []string{
		"junod", "tx", "authz", "exec", node.HomeDir() + "/" + filePath,
		"--home", chain.HomeDir(),
		"--node", chain.GetRPCAddress(),
		"--from", grantee.KeyName(),
		"--chain-id", chain.Config().ChainID,
		"--gas", "500000",
		"--fees", feeCoin,
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	}

	if amount != "" {
		cmd = append(cmd, "--amount", amount)
	}

	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	debugOutput(t, string(stdout))

	if err := testutil.WaitForBlocks(ctx, 2, chain); err != nil {
		t.Fatal(err)
	}
}
