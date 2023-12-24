package helpers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

func SmartQueryString(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, contractAddr, queryMsg string, res interface{}) error {
	var jsonMap map[string]interface{}
	if err := json.Unmarshal([]byte(queryMsg), &jsonMap); err != nil {
		t.Fatal(err)
	}
	err := chain.QueryContract(ctx, contractAddr, jsonMap, &res)
	return err
}

func StoreContract(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, keyname string, fileLoc string) (codeId string) {
	codeId, err := chain.StoreContract(ctx, keyname, fileLoc)
	if err != nil {
		t.Fatal(err)
	}
	return codeId
}

func SetupContract(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, keyname string, fileLoc string, message string, extraFlags ...string) (codeId, contract string) {
	codeId = StoreContract(t, ctx, chain, keyname, fileLoc)

	needsNoAdminFlag := true
	// if extraFlags contains "--admin", switch to false
	for _, flag := range extraFlags {
		if flag == "--admin" {
			needsNoAdminFlag = false
		}
	}

	contractAddr, err := chain.InstantiateContract(ctx, keyname, codeId, message, needsNoAdminFlag, extraFlags...)
	if err != nil {
		t.Fatal(err)
	}

	return codeId, contractAddr
}

func MigrateContract(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, keyname string, contractAddr string, fileLoc string, message string) (codeId, contract string) {
	codeId, err := chain.StoreContract(ctx, keyname, fileLoc)
	if err != nil {
		t.Fatal(err)
	}

	// Execute migrate tx
	cmd := []string{
		"junod", "tx", "wasm", "migrate", contractAddr, codeId, message,
		"--node", chain.GetRPCAddress(),
		"--home", chain.HomeDir(),
		"--chain-id", chain.Config().ChainID,
		"--from", keyname,
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

	return codeId, contractAddr
}

func ExecuteMsgWithAmount(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, contractAddr, amount, message string) {
	// amount is #utoken

	// There has to be a way to do this in ictest?
	cmd := []string{
		"junod", "tx", "wasm", "execute", contractAddr, message,
		"--node", chain.GetRPCAddress(),
		"--home", chain.HomeDir(),
		"--chain-id", chain.Config().ChainID,
		"--from", user.KeyName(),
		"--gas", "500000",
		"--amount", amount,
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

func ExecuteMsgWithFee(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, contractAddr, amount, feeCoin, message string) {
	// amount is #utoken

	// There has to be a way to do this in ictest?
	cmd := []string{
		"junod", "tx", "wasm", "execute", contractAddr, message,
		"--node", chain.GetRPCAddress(),
		"--home", chain.HomeDir(),
		"--chain-id", chain.Config().ChainID,
		"--from", user.KeyName(),
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

func ExecuteMsgWithFeeReturn(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, contractAddr, amount, feeCoin, message string) (*sdk.TxResponse, error) {
	// amount is #utoken

	// There has to be a way to do this in ictest? (there is, use node.ExecTx)
	cmd := []string{
		"wasm", "execute", contractAddr, message,
		"--output", "json",
		"--node", chain.GetRPCAddress(),
		"--home", chain.HomeDir(),
		"--gas", "500000",
		"--fees", feeCoin,
		"--keyring-dir", chain.HomeDir(),
	}

	if amount != "" {
		cmd = append(cmd, "--amount", amount)
	}

	node := chain.GetNode()

	txHash, err := node.ExecTx(ctx, user.KeyName(), cmd...)
	if err != nil {
		return nil, err
	}

	// convert stdout into a TxResponse
	txRes, err := chain.GetTransaction(txHash)
	return txRes, err
}
