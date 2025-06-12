package bank

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// KeeperInterface defines the interface for the keeper
type KeeperInterface interface {
	GetQueryContext() (context.Context, error)
	ValidateDenom(ctx context.Context, denom string) error
	GetBankKeeper() BankKeeperInterface
}

// BankKeeperInterface defines the interface for the bank keeper
type BankKeeperInterface interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}
