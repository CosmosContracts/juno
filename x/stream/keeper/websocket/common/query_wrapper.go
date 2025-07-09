package common

import (
	"context"
	"net/http"
)

// QueryBuilder helps build query functions with less boilerplate
type QueryBuilder struct {
	provider QueryContextProvider
	logger   Logger
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(provider QueryContextProvider, logger Logger) *QueryBuilder {
	return &QueryBuilder{
		provider: provider,
		logger:   logger,
	}
}

// WrapSimpleQuery wraps a query that returns data without error
func (qb *QueryBuilder) WrapSimpleQuery(queryFunc func(context.Context) any) func() any {
	return func() any {
		ctx, err := qb.provider.GetQueryContext()
		if err != nil {
			qb.logger.Error("failed to get query context", "error", err)
			return ErrorResponse("service temporarily unavailable")
		}
		return queryFunc(ctx)
	}
}

// WrapQueryWithError wraps a query that returns data and error
func (qb *QueryBuilder) WrapQueryWithError(queryFunc func(context.Context) (any, error)) func() any {
	return func() any {
		ctx, err := qb.provider.GetQueryContext()
		if err != nil {
			qb.logger.Error("failed to get query context", "error", err)
			return ErrorResponse("service temporarily unavailable")
		}

		data, err := queryFunc(ctx)
		if err != nil {
			qb.logger.Error("query function error", "error", err)
			return ErrorResponse("failed to retrieve data")
		}
		return data
	}
}

// StandardSubscriptionParams helps build ConnectionParams with less boilerplate
type StandardSubscriptionParams struct {
	Request         *http.Request
	ValidationFunc  func() error
	SubscriptionKey SubscriptionKey
	InitialQuery    func(context.Context) (any, error)
	UpdateQuery     func(context.Context) any
}

// BuildConnectionParams builds standard connection parameters
func (qb *QueryBuilder) BuildConnectionParams(w http.ResponseWriter, params StandardSubscriptionParams) ConnectionParams {
	return ConnectionParams{
		Writer:          w,
		Request:         params.Request,
		ValidationFunc:  params.ValidationFunc,
		InitialDataFunc: params.InitialQuery,
		SubscriptionKey: params.SubscriptionKey,
		QueryFunc:       qb.WrapSimpleQuery(params.UpdateQuery),
	}
}

// CreateSimpleHandler creates a standard WebSocket handler with minimal boilerplate
func CreateSimpleHandler(handler *ModuleHandler, params StandardSubscriptionParams) RouteHandler {
	return func(w http.ResponseWriter, _ *http.Request) {
		qb := NewQueryBuilder(handler, handler.Logger)
		connParams := qb.BuildConnectionParams(w, params)
		handler.HandleStandardConnection(connParams)
	}
}
