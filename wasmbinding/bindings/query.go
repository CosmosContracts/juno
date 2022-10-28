package bindings

import (
	oracle "github.com/CosmosContracts/juno/v11/x/oracle/wasm"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// DenomQuery contains denom custom queries.

type CosmosQuery struct {
	Bank   *BankQuery
	Oracle *oracle.OracleQuery
}

type BankQuery struct {
	DenomMetadata *banktypes.QueryDenomMetadataRequest `json:"denom_metadata,omitempty"`
	Supply        *banktypes.QuerySupplyOfRequest      `json:"supply,omitempty"`
}
