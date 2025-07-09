package websocket

import (
	"context"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/bank"
	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/common"
	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/staking"
)

// KeeperInterface defines what the websocket package needs from the main keeper
type KeeperInterface interface {
	GetQueryContext() (context.Context, error)
	ValidateDenom(ctx context.Context, denom string) error
	GetBankKeeper() bankkeeper.Keeper
	GetStakingKeeper() stakingkeeper.Keeper
}

// CreateHandler creates a new WebSocket handler from a keeper that implements KeeperInterface
func CreateHandler(
	ctx context.Context,
	keeper KeeperInterface,
	config *common.StreamConfig,
	logger common.Logger,
	connManager common.ConnectionManager,
	registry common.SubscriptionRegistry,
	circuitBreaker common.CircuitBreaker,
	allowAllOrigins bool,
) *Handler {
	// Create adapters for bank and staking
	bankAdapter := &bankKeeperAdapter{keeper: keeper}
	stakingAdapter := &stakingKeeperAdapter{keeper: keeper}

	return NewHandler(
		ctx,
		bankAdapter,
		stakingAdapter,
		config,
		logger,
		connManager,
		registry,
		circuitBreaker,
		allowAllOrigins,
	)
}

// bankKeeperAdapter implements bank.KeeperInterface
var _ bank.KeeperInterface = (*bankKeeperAdapter)(nil)

type bankKeeperAdapter struct {
	keeper KeeperInterface
}

func (a *bankKeeperAdapter) GetQueryContext() (context.Context, error) {
	return a.keeper.GetQueryContext()
}

func (a *bankKeeperAdapter) ValidateDenom(ctx context.Context, denom string) error {
	return a.keeper.ValidateDenom(ctx, denom)
}

func (a *bankKeeperAdapter) GetBankQueryServer() banktypes.QueryServer {
	// The bank keeper implements the QueryServer interface
	return a.keeper.GetBankKeeper()
}

// stakingKeeperAdapter implements staking.KeeperInterface
var _ staking.KeeperInterface = (*stakingKeeperAdapter)(nil)

type stakingKeeperAdapter struct {
	keeper KeeperInterface
}

func (a *stakingKeeperAdapter) GetQueryContext() (context.Context, error) {
	return a.keeper.GetQueryContext()
}

func (a *stakingKeeperAdapter) GetStakingQueryServer() stakingtypes.QueryServer {
	// The staking keeper needs to be wrapped in a Querier to implement QueryServer
	sk := a.keeper.GetStakingKeeper()
	return stakingkeeper.NewQuerier(&sk)
}
