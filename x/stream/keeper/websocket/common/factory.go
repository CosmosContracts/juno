package common

import (
	"context"

	"github.com/gorilla/websocket"
)

// HandlerDependencies contains all common dependencies needed by handlers
type HandlerDependencies struct {
	Config         *StreamConfig
	Logger         Logger
	ConnManager    ConnectionManager
	Registry       SubscriptionRegistry
	CircuitBreaker CircuitBreaker
	AppContext     context.Context
	Upgrader       *websocket.Upgrader
}

// NewHandlerDependencies creates a new HandlerDependencies with all required components
func NewHandlerDependencies(
	appContext context.Context,
	config *StreamConfig,
	logger Logger,
	connManager ConnectionManager,
	registry SubscriptionRegistry,
	circuitBreaker CircuitBreaker,
	allowAllOrigins bool,
) *HandlerDependencies {
	return &HandlerDependencies{
		Config:         config,
		Logger:         logger,
		ConnManager:    connManager,
		Registry:       registry,
		CircuitBreaker: circuitBreaker,
		AppContext:     appContext,
		Upgrader:       GetUpgrader(allowAllOrigins),
	}
}

// CreateBaseHandler creates a BaseHandler from dependencies
func (d *HandlerDependencies) CreateBaseHandler() BaseHandler {
	return BaseHandler{
		Config:         d.Config,
		Logger:         d.Logger,
		ConnManager:    d.ConnManager,
		Registry:       d.Registry,
		CircuitBreaker: d.CircuitBreaker,
		AppContext:     d.AppContext,
		Upgrader:       d.Upgrader,
	}
}

// QueryContextProvider is a common interface for keepers that provide query context
type QueryContextProvider interface {
	GetQueryContext() (context.Context, error)
}
