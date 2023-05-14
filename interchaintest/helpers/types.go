package helpers

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Go based data types for querying on the contract.

// TODO: Auto generate in the future from Rust types -> Go types?
// Execute types are not needed here. We just use strings. Could add though in the future and to_string it

// EntryPoint
type QueryMsg struct {
	GetConfig *struct{} `json:"get_config,omitempty"`

	GetBalance     *GetBalanceQuery     `json:"get_balance,omitempty"`
	GetAllBalances *GetAllBalancesQuery `json:"get_all_balances,omitempty"`
}

type GetAllBalancesQuery struct {
	Address string `json:"address"`
}
type GetAllBalancesResponse struct {
	// or is it wasm Coin type?
	Data []sdk.Coin `json:"data"`
}

// {"get_balance":{"address":"juno1...","denom":"factory/juno1.../RcqfWz"}}
type GetBalanceQuery struct {
	Address string `json:"address"`
	Denom   string `json:"denom"`
}
type GetBalanceResponse struct {
	// or is it wasm Coin type?
	Data sdk.Coin `json:"data"`
}
