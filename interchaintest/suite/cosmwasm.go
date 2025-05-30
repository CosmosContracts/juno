package suite

import (
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *E2ETestSuite) SmartQueryString(chain *cosmos.CosmosChain, contractAddr, queryMsg string, res any) error {
	t := s.T()
	var jsonMap map[string]any
	if err := json.Unmarshal([]byte(queryMsg), &jsonMap); err != nil {
		t.Fatal(err)
	}
	err := chain.QueryContract(s.Ctx, contractAddr, jsonMap, &res)
	return err
}

func (s *E2ETestSuite) StoreContract(chain *cosmos.CosmosChain, keyName string, fileLoc string, fees sdk.Coins) (codeId string) {
	cn := chain.GetNode()

	_, file := filepath.Split(fileLoc)
	err := cn.CopyFile(s.Ctx, fileLoc, file)
	if err != nil {
		s.T().Fatal(fmt.Errorf("writing contract file to docker volume: %w", err))
	}

	_, err = s.ExecTx(
		chain,
		keyName,
		false,
		false,
		"wasm",
		"store",
		path.Join(cn.HomeDir(), file),
		"--fees",
		fees.String(),
		"--gas",
		"auto",
	)
	if err != nil {
		s.T().Fatal(err)
	}

	stdout, _, err := cn.ExecQuery(s.Ctx, "wasm", "list-code", "--reverse")
	if err != nil {
		s.T().Fatal(err)
	}

	res := cosmos.CodeInfosResponse{}
	if err := json.Unmarshal(stdout, &res); err != nil {
		s.T().Fatal(err)
	}

	return res.CodeInfos[0].CodeID
}

func (s *E2ETestSuite) SetupContract(chain *cosmos.CosmosChain, keyname string, fileLoc string, initMessage string, skipTxCheck bool, fees sdk.Coins, extraFlags ...string) (codeId, contract string) {
	t := s.T()

	codeId = s.StoreContract(chain, keyname, fileLoc, fees)

	needsNoAdminFlag := true
	// if extraFlags contains "--admin", switch to false
	for _, flag := range extraFlags {
		if flag == "--admin" {
			needsNoAdminFlag = false
		}
	}

	contractAddr, err := s.InstantiateContract(chain, keyname, codeId, initMessage, fees, needsNoAdminFlag, skipTxCheck, extraFlags...)
	if err != nil {
		t.Fatal(err)
	}

	return codeId, contractAddr
}

func (s *E2ETestSuite) MigrateContract(chain *cosmos.CosmosChain, keyName string, contractAddr string, fileLoc string, message string, fees sdk.Coins) (codeId, contract string) {
	t := s.T()
	codeId = s.StoreContract(s.Chain, keyName, fileLoc, fees)

	// Execute migrate tx
	txHash, err := s.ExecTx(
		chain,
		keyName,
		false,
		false,
		"wasm",
		"migrate",
		contractAddr, codeId, message,
		"--fees",
		fees.String(),
		"--gas", "auto",
	)
	if err != nil {
		t.Fatal(err)
	}

	// convert stdout into a TxResponse
	txRes, err := chain.GetTransaction(txHash)
	if err != nil {
		t.Fatal(err)
	}
	s.DebugOutput(string(txRes.RawLog))

	return codeId, contractAddr
}

// InstantiateContract takes a code id for a smart contract and initialization message and returns the instantiated contract address.
func (s *E2ETestSuite) InstantiateContract(chain *cosmos.CosmosChain, keyName string, codeID string, initMessage string, fees sdk.Coins, needsNoAdminFlag, skipTxCheck bool, extraExecTxArgs ...string) (string, error) {
	command := []string{
		"wasm", "instantiate", codeID, initMessage, "--label", "wasm-contract",
		"--fees", fees.String(),
		"--gas", "auto",
	}
	command = append(command, extraExecTxArgs...)
	if needsNoAdminFlag {
		command = append(command, "--no-admin")
	}
	txHash, err := s.ExecTx(chain, keyName, false, skipTxCheck, command...)
	if err != nil {
		return "", err
	}

	tn := chain.GetNode()

	txResp, err := tn.GetTransaction(tn.CliContext(), txHash)
	if err != nil {
		return "", fmt.Errorf("failed to get transaction %s: %w", txHash, err)
	}
	if txResp.Code != 0 {
		return "", fmt.Errorf("error in transaction (code: %d): %s", txResp.Code, txResp.RawLog)
	}

	stdout, _, err := tn.ExecQuery(s.Ctx, "wasm", "list-contract-by-code", codeID)
	if err != nil {
		return "", err
	}

	contactsRes := cosmos.QueryContractResponse{}
	if err := json.Unmarshal(stdout, &contactsRes); err != nil {
		return "", err
	}

	contractAddress := contactsRes.Contracts[len(contactsRes.Contracts)-1]
	return contractAddress, nil
}

func (s *E2ETestSuite) ExecuteMsgWithAmount(chain *cosmos.CosmosChain, user ibc.Wallet, contractAddr, amount, message string, fees sdk.Coins) (*sdk.TxResponse, error) {
	t := s.T()

	cmd := []string{
		"wasm", "execute", contractAddr, message,
		"--from", user.KeyName(),
		"--gas", "auto",
		"--fees", fees.String(),
		"--amount", amount,
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

	if err := testutil.WaitForBlocks(s.Ctx, 2, chain); err != nil {
		t.Fatal(err)
	}

	return txRes, err
}

func (s *E2ETestSuite) ExecuteMsgWithFeeReturn(chain *cosmos.CosmosChain, user ibc.Wallet, contractAddr, amount, message string, skipTxCheck bool, fees sdk.Coins) (*sdk.TxResponse, error) {
	t := s.T()
	cmd := []string{
		"wasm", "execute", contractAddr, message,
		"--gas", "auto",
		"--fees", fees.String(),
	}

	if amount != "" {
		cmd = append(cmd, "--amount", amount)
	}

	txHash, _ := s.ExecTx(s.Chain, user.KeyName(), false, skipTxCheck, cmd...)
	if skipTxCheck {
		if err := testutil.WaitForBlocks(s.Ctx, 1, chain); err != nil {
			t.Fatal(err)
		}
		txRes, _ := chain.GetTransaction(txHash)
		return txRes, nil
	}
	if err := testutil.WaitForBlocks(s.Ctx, 1, chain); err != nil {
		t.Fatal(err)
	}
	// convert stdout into a TxResponse
	txRes, err := chain.GetTransaction(txHash)
	return txRes, err
}
