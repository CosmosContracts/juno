package bank

import (
	"context"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// KeeperInterface defines the interface for the keeper
type KeeperInterface interface {
	GetQueryContext() (context.Context, error)
	ValidateDenom(ctx context.Context, denom string) error
	GetBankQueryServer() banktypes.QueryServer
}
