package suite

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Go based data types for querying on the contract.
// Execute types are not needed here. We just use strings. Could add though in the future and to_string it

// EntryPoint
type ContractQueryMsg struct {
	// Tokenfactory Core
	GetConfig      *struct{}            `json:"get_config,omitempty"`
	GetBalance     *GetBalanceQuery     `json:"get_balance,omitempty"`
	GetAllBalances *GetAllBalancesQuery `json:"get_all_balances,omitempty"`

	// Unity Contract
	GetWithdrawalReadyTime *struct{} `json:"get_withdrawal_ready_time,omitempty"`

	// IBCHooks
	GetCount      *GetCountQuery      `json:"get_count,omitempty"`
	GetTotalFunds *GetTotalFundsQuery `json:"get_total_funds,omitempty"`
}

type GetAllBalancesQuery struct {
	Address string `json:"address"`
}
type GetAllBalancesResponse struct {
	Data []sdk.Coin `json:"data"`
}

type GetBalanceQuery struct {
	Address string `json:"address"`
	Denom   string `json:"denom"`
}
type GetBalanceResponse struct {
	Data sdk.Coin `json:"data"`
}

type WithdrawalTimestampResponse struct {
	Data *WithdrawalTimestampObj `json:"data"`
}
type WithdrawalTimestampObj struct {
	WithdrawalReadyTimestamp string `json:"withdrawal_ready_timestamp"`
}

type GetTotalFundsQuery struct {
	Addr string `json:"addr"`
}
type GetTotalFundsResponse struct {
	Data *GetTotalFundsObj `json:"data"`
}
type GetTotalFundsObj struct {
	TotalFunds []WasmCoin `json:"total_funds"`
}

type WasmCoin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type GetCountQuery struct {
	Addr string `json:"addr"`
}
type GetCountResponse struct {
	Data *GetCountObj `json:"data"`
}
type GetCountObj struct {
	Count int64 `json:"count"`
}

type ClockContractResponse struct {
	Data *ClockContractObj `json:"data"`
}
type ClockContractObj struct {
	Val uint32 `json:"val"`
}

type GetCwHooksDelegationResponse struct {
	Data *GetDelegationObj `json:"data"`
}
type GetDelegationObj struct {
	ValidatorAddress string `json:"validator_address"`
	DelegatorAddress string `json:"delegator_address"`
	Shares           string `json:"shares"`
}

type ClockContract struct {
	ClockContract struct {
		ContractAddress string `json:"contract_address"`
		IsJailed        bool   `json:"is_jailed"`
	} `json:"clock_contract"`
}
