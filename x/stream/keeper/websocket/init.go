package websocket

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/bank"
	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/common"
	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/staking"
)

// KeeperInterface defines what the websocket package needs from the main keeper
type KeeperInterface interface {
	GetQueryContext() (context.Context, error)
	ValidateDenom(ctx context.Context, denom string) error
	GetBankKeeper() BankKeeper
	GetStakingKeeper() StakingKeeper
}

// BankKeeper interface for bank operations
type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}

// StakingKeeper interface for staking operations
type StakingKeeper interface {
	GetAllDelegatorDelegations(ctx context.Context, delegator sdk.AccAddress) ([]stakingtypes.Delegation, error)
	GetDelegation(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (stakingtypes.Delegation, error)
	GetValidator(ctx context.Context, addr sdk.ValAddress) (stakingtypes.Validator, error)
	BondDenom(ctx context.Context) (string, error)
	GetAllUnbondingDelegations(ctx context.Context, delegator sdk.AccAddress) ([]stakingtypes.UnbondingDelegation, error)
	GetUnbondingDelegation(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (stakingtypes.UnbondingDelegation, error)
}

// CreateHandler creates a new WebSocket handler from a keeper that implements KeeperInterface
func CreateHandler(
	keeper KeeperInterface,
	config *common.StreamConfig,
	logger common.Logger,
	connManager common.ConnectionManager,
	registry common.SubscriptionRegistry,
	circuitBreaker common.CircuitBreaker,
	appContext context.Context,
	allowAllOrigins bool,
) *Handler {
	// Create adapters for bank and staking
	bankAdapter := &bankKeeperAdapter{keeper: keeper}
	stakingAdapter := &stakingKeeperAdapter{keeper: keeper}

	return NewHandler(
		bankAdapter,
		stakingAdapter,
		config,
		logger,
		connManager,
		registry,
		circuitBreaker,
		appContext,
		allowAllOrigins,
	)
}

// bankKeeperAdapter implements bank.KeeperInterface
type bankKeeperAdapter struct {
	keeper KeeperInterface
}

func (a *bankKeeperAdapter) GetQueryContext() (context.Context, error) {
	return a.keeper.GetQueryContext()
}

func (a *bankKeeperAdapter) ValidateDenom(ctx context.Context, denom string) error {
	return a.keeper.ValidateDenom(ctx, denom)
}

func (a *bankKeeperAdapter) GetBankKeeper() bank.BankKeeperInterface {
	return a.keeper.GetBankKeeper()
}

// stakingKeeperAdapter implements staking.KeeperInterface
type stakingKeeperAdapter struct {
	keeper KeeperInterface
}

func (a *stakingKeeperAdapter) GetQueryContext() (context.Context, error) {
	return a.keeper.GetQueryContext()
}

func (a *stakingKeeperAdapter) GetStakingKeeper() staking.StakingKeeperInterface {
	return a.keeper.GetStakingKeeper()
}