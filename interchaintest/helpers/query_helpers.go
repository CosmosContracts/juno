package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/stretchr/testify/require"
)

func GetUserTokenFactoryBalances(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, contract string, uaddr string) GetAllBalancesResponse {
	var res GetAllBalancesResponse
	err := chain.QueryContract(ctx, contract, QueryMsg{GetAllBalances: &GetAllBalancesQuery{Address: uaddr}}, &res)
	require.NoError(t, err)
	return res
}

func GetUnityContractWithdrawalReadyTime(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, contract string) WithdrawalTimestampResponse {
	// junod query wasm contract-state smart <contract> '{"get_withdrawal_ready_time":{}}' --output json
	var res WithdrawalTimestampResponse
	err := chain.QueryContract(ctx, contract, QueryMsg{GetWithdrawalReadyTime: &struct{}{}}, &res)
	require.NoError(t, err)
	return res
}

func GetClockContractValue(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, contract string) ClockContractResponse {
	var res ClockContractResponse
	err := chain.QueryContract(ctx, contract, QueryMsg{GetConfig: &struct{}{}}, &res)
	require.NoError(t, err)
	return res
}

// From stakingtypes.Validator
type Vals struct {
	Validators []struct {
		OperatorAddress string `json:"operator_address"`
		ConsensusPubkey struct {
			Type string `json:"@type"`
			Key  string `json:"key"`
		} `json:"consensus_pubkey"`
		Jailed          bool   `json:"jailed"`
		Status          string `json:"status"`
		Tokens          string `json:"tokens"`
		DelegatorShares string `json:"delegator_shares"`
		Description     struct {
			Moniker         string `json:"moniker"`
			Identity        string `json:"identity"`
			Website         string `json:"website"`
			SecurityContact string `json:"security_contact"`
			Details         string `json:"details"`
		} `json:"description"`
		UnbondingHeight string    `json:"unbonding_height"`
		UnbondingTime   time.Time `json:"unbonding_time"`
		Commission      struct {
			CommissionRates struct {
				Rate          string `json:"rate"`
				MaxRate       string `json:"max_rate"`
				MaxChangeRate string `json:"max_change_rate"`
			} `json:"commission_rates"`
			UpdateTime time.Time `json:"update_time"`
		} `json:"commission"`
		MinSelfDelegation       string `json:"min_self_delegation"`
		UnbondingOnHoldRefCount string `json:"unbonding_on_hold_ref_count"`
		UnbondingIds            []any  `json:"unbonding_ids"`
	} `json:"validators"`
	Pagination struct {
		NextKey any    `json:"next_key"`
		Total   string `json:"total"`
	} `json:"pagination"`
}

func GetValidators(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain) Vals {
	var res Vals

	cmd := []string{
		"junod", "query", "staking", "validators",
		"--output", "json",
		"--node", chain.GetRPCAddress(),
	}

	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	// print stdout
	fmt.Println(string(stdout))

	// put the stdout json into res
	if err := json.Unmarshal(stdout, &res); err != nil {
		t.Fatal(err)
	}

	return res
}

// FeePay

type FeePayUses struct {
	Uses string `json:"uses"`
}

func GetFeePayUses(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, contract, wallet string) FeePayUses {
	var res FeePayUses

	cmd := []string{
		"junod", "query", "feepay", "uses", contract, wallet,
		"--output", "json",
		"--node", chain.GetRPCAddress(),
	}

	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	fmt.Println(string(stdout))

	if err := json.Unmarshal(stdout, &res); err != nil {
		t.Fatal(err)
	}

	return res
}

type FeePayContracts struct {
	FeePayContracts []struct {
		ContractAddress string `json:"contract_address"`
		Balance         string `json:"balance"`
		WalletLimit     string `json:"wallet_limit"`
	} `json:"fee_pay_contracts"`
	Pagination struct {
		NextKey any    `json:"next_key"`
		Total   string `json:"total"`
	} `json:"pagination"`
}

func GetFeePayContracts(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain) FeePayContracts {
	var res FeePayContracts

	cmd := []string{
		"junod", "query", "feepay", "contracts",
		"--output", "json",
		"--node", chain.GetRPCAddress(),
	}

	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	fmt.Println(string(stdout))

	if err := json.Unmarshal(stdout, &res); err != nil {
		t.Fatal(err)
	}

	return res
}

type FeePayContract struct {
	FeePayContract struct {
		ContractAddress string `json:"contract_address"`
		Balance         string `json:"balance"`
		WalletLimit     string `json:"wallet_limit"`
	} `json:"fee_pay_contract"`
}

func GetFeePayContract(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, contract string) FeePayContract {
	var res FeePayContract

	cmd := []string{
		"junod", "query", "feepay", "contract", contract,
		"--output", "json",
		"--node", chain.GetRPCAddress(),
	}

	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	fmt.Println(string(stdout))

	if err := json.Unmarshal(stdout, &res); err != nil {
		t.Fatal(err)
	}

	return res
}
