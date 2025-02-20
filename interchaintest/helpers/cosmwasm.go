package helpers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
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
		"--home", chain.HomeDir(),
		"--node", chain.GetRPCAddress(),
		"--from", keyname,
		"--chain-id", chain.Config().ChainID,
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

func ExecuteMsgWithAmount(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, contractAddr, amount, message string) (*sdk.TxResponse, error) {
	// amount is #utoken

	// There has to be a way to do this in ictest?
	cmd := []string{
		"wasm", "execute", contractAddr, message,
		"--home", chain.HomeDir(),
		"--from", user.KeyName(),
		"--gas", "500000",
		"--amount", amount,
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	}
	node := chain.GetNode()
	txHash, err := node.ExecTx(ctx, user.KeyName(), cmd...)
	if err != nil {
		t.Fatal(err)
	}
	// convert stdout into a TxResponse
	txRes, err := chain.GetTransaction(txHash)
	if err != nil {
		t.Fatal(err)
	}

	if err := testutil.WaitForBlocks(ctx, 2, chain); err != nil {
		t.Fatal(err)
	}

	return txRes, err
}

func ExecuteMsgWithFeeReturn(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, contractAddr, amount, feeCoin, message string) (*sdk.TxResponse, error) {
	// amount is #utoken (e.g. "1000ujuno")

	// There has to be a way to do this in ictest? (there is, use node.ExecTx)
	cmd := []string{
		"wasm", "execute", contractAddr, message,
		"--output", "json",
		"--home", chain.HomeDir(),
		"--gas", "500000",
		"--fees", feeCoin,
		"--keyring-dir", chain.HomeDir(),
	}

	if amount != "" {
		cmd = append(cmd, "--amount", amount)
	}

	node := chain.GetNode()

	txHash, _ := node.ExecTx(ctx, user.KeyName(), cmd...)
	// convert stdout into a TxResponse
	txRes, err := chain.GetTransaction(txHash)
	return txRes, err
}
