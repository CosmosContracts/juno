package staking

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/types/query"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/common"
	"github.com/CosmosContracts/juno/v30/x/stream/types"
)

// Module implements the staking WebSocket module
type Module struct {
	*common.ModuleHandler
	keeper           KeeperInterface
	name             string
	routeDefinitions []common.RouteDefinition
}

// NewModule creates a new staking module
func NewModule(keeper KeeperInterface, deps *common.HandlerDependencies) *Module {
	m := &Module{
		ModuleHandler: common.NewModuleHandler(deps, keeper),
		keeper:        keeper,
		name:          "staking",
	}

	// Register routes
	m.RegisterRoute("delegations", m.handleDelegations)
	m.RegisterRoute("delegation", m.handleDelegation)
	m.RegisterRoute("unbonding_delegations", m.handleUnbondingDelegations)
	m.RegisterRoute("unbonding_delegation", m.handleUnbondingDelegation)

	// Define route patterns for automatic registration
	m.routeDefinitions = []common.RouteDefinition{
		{Pattern: "/ws/subscribe/staking/delegations/{delegator}", Handler: m.handleDelegations},
		{Pattern: "/ws/subscribe/staking/delegation/{delegator}/{validator}", Handler: m.handleDelegation},
		{Pattern: "/ws/subscribe/staking/unbonding-delegations/{delegator}", Handler: m.handleUnbondingDelegations},
		{Pattern: "/ws/subscribe/staking/unbonding-delegation/{delegator}/{validator}", Handler: m.handleUnbondingDelegation},
	}

	return m
}

// Name returns the module name
func (m *Module) Name() string {
	return m.name
}

// Routes returns all routes for this module
func (m *Module) Routes() map[string]common.RouteHandler {
	return m.GetRoutes()
}

// RouteDefinitions returns all route definitions with full patterns
func (m *Module) RouteDefinitions() []common.RouteDefinition {
	return m.routeDefinitions
}

// handleDelegations handles delegations subscription requests
func (m *Module) handleDelegations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	delegatorAddress := vars["delegator"]

	// Validate address
	addrValidator := common.ValidateAccAddress(delegatorAddress)
	if !addrValidator.IsValid() {
		http.Error(w, addrValidator.Error().Error(), http.StatusBadRequest)
		return
	}

	qb := common.NewQueryBuilder(m.keeper, m.Logger)

	params := common.StandardSubscriptionParams{
		Request:        r,
		ValidationFunc: addrValidator.ValidationFunc(),
		SubscriptionKey: common.NewSubscriptionKeyAdapter(
			types.GenerateSubscriptionKey(types.SubscriptionTypeDelegations, delegatorAddress, "", ""),
		),
		InitialQuery: func(ctx context.Context) (any, error) {
			req := &stakingtypes.QueryDelegatorDelegationsRequest{
				DelegatorAddr: delegatorAddress,
				Pagination: &query.PageRequest{
					Limit: 1000,
				},
			}
			resp, err := m.keeper.GetStakingQueryServer().DelegatorDelegations(ctx, req)
			if err != nil {
				return nil, err
			}
			// Convert to our response format
			var responses []stakingtypes.DelegationResponse
			for _, del := range resp.DelegationResponses {
				responses = append(responses, stakingtypes.DelegationResponse{
					Delegation: del.Delegation,
					Balance:    del.Balance,
				})
			}
			return responses, nil
		},
		UpdateQuery: func(ctx context.Context) any {
			req := &stakingtypes.QueryDelegatorDelegationsRequest{
				DelegatorAddr: delegatorAddress,
				Pagination: &query.PageRequest{
					Limit: 1000,
				},
			}
			resp, err := m.keeper.GetStakingQueryServer().DelegatorDelegations(ctx, req)
			if err != nil {
				return nil
			}
			var responses []stakingtypes.DelegationResponse
			for _, del := range resp.DelegationResponses {
				responses = append(responses, stakingtypes.DelegationResponse{
					Delegation: del.Delegation,
					Balance:    del.Balance,
				})
			}
			return responses
		},
	}

	connParams := qb.BuildConnectionParams(w, params)
	m.HandleStandardConnection(connParams)
}

// handleDelegation handles single delegation subscription requests
func (m *Module) handleDelegation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	delegatorAddress := vars["delegator"]
	validatorAddress := vars["validator"]

	// Validate addresses
	delAddrValidator := common.ValidateAccAddress(delegatorAddress)
	valAddrValidator := common.ValidateValAddress(validatorAddress)

	validator := common.NewCompositeValidator().
		AddAddressValidator(delAddrValidator).
		AddAddressValidator(valAddrValidator)

	if err := validator.ValidationFunc()(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	qb := common.NewQueryBuilder(m.keeper, m.Logger)

	params := common.StandardSubscriptionParams{
		Request:        r,
		ValidationFunc: validator.ValidationFunc(),
		SubscriptionKey: common.NewSubscriptionKeyAdapter(
			types.GenerateSubscriptionKey(types.SubscriptionTypeDelegation, delegatorAddress, validatorAddress, ""),
		),
		InitialQuery: func(ctx context.Context) (any, error) {
			req := &stakingtypes.QueryDelegationRequest{
				DelegatorAddr: delegatorAddress,
				ValidatorAddr: validatorAddress,
			}
			resp, err := m.keeper.GetStakingQueryServer().Delegation(ctx, req)
			if err != nil {
				return nil, err
			}
			if resp.DelegationResponse == nil {
				return nil, nil
			}
			return resp.DelegationResponse, nil
		},
		UpdateQuery: func(ctx context.Context) any {
			req := &stakingtypes.QueryDelegationRequest{
				DelegatorAddr: delegatorAddress,
				ValidatorAddr: validatorAddress,
			}
			resp, err := m.keeper.GetStakingQueryServer().Delegation(ctx, req)
			if err != nil {
				return nil
			}
			if resp.DelegationResponse == nil {
				return nil
			}
			return resp.DelegationResponse
		},
	}

	connParams := qb.BuildConnectionParams(w, params)
	m.HandleStandardConnection(connParams)
}

// handleUnbondingDelegations handles unbonding delegations subscription requests
func (m *Module) handleUnbondingDelegations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	delegatorAddress := vars["delegator"]

	// Validate address
	addrValidator := common.ValidateAccAddress(delegatorAddress)
	if !addrValidator.IsValid() {
		http.Error(w, addrValidator.Error().Error(), http.StatusBadRequest)
		return
	}

	qb := common.NewQueryBuilder(m.keeper, m.Logger)

	params := common.StandardSubscriptionParams{
		Request:        r,
		ValidationFunc: addrValidator.ValidationFunc(),
		SubscriptionKey: common.NewSubscriptionKeyAdapter(
			types.GenerateSubscriptionKey(types.SubscriptionTypeUnbondingDelegations, delegatorAddress, "", ""),
		),
		InitialQuery: func(ctx context.Context) (any, error) {
			req := &stakingtypes.QueryDelegatorUnbondingDelegationsRequest{
				DelegatorAddr: delegatorAddress,
				Pagination: &query.PageRequest{
					Limit: 1000,
				},
			}
			resp, err := m.keeper.GetStakingQueryServer().DelegatorUnbondingDelegations(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.UnbondingResponses, nil
		},
		UpdateQuery: func(ctx context.Context) any {
			req := &stakingtypes.QueryDelegatorUnbondingDelegationsRequest{
				DelegatorAddr: delegatorAddress,
				Pagination: &query.PageRequest{
					Limit: 1000,
				},
			}
			resp, err := m.keeper.GetStakingQueryServer().DelegatorUnbondingDelegations(ctx, req)
			if err != nil {
				return nil
			}
			return resp.UnbondingResponses
		},
	}

	connParams := qb.BuildConnectionParams(w, params)
	m.HandleStandardConnection(connParams)
}

// handleUnbondingDelegation handles single unbonding delegation subscription requests
func (m *Module) handleUnbondingDelegation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	delegatorAddress := vars["delegator"]
	validatorAddress := vars["validator"]

	// Validate addresses
	delAddrValidator := common.ValidateAccAddress(delegatorAddress)
	valAddrValidator := common.ValidateValAddress(validatorAddress)

	validator := common.NewCompositeValidator().
		AddAddressValidator(delAddrValidator).
		AddAddressValidator(valAddrValidator)

	if err := validator.ValidationFunc()(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	qb := common.NewQueryBuilder(m.keeper, m.Logger)

	params := common.StandardSubscriptionParams{
		Request:        r,
		ValidationFunc: validator.ValidationFunc(),
		SubscriptionKey: common.NewSubscriptionKeyAdapter(
			types.GenerateSubscriptionKey(types.SubscriptionTypeUnbondingDelegation, delegatorAddress, validatorAddress, ""),
		),
		InitialQuery: func(ctx context.Context) (any, error) {
			req := &stakingtypes.QueryUnbondingDelegationRequest{
				DelegatorAddr: delegatorAddress,
				ValidatorAddr: validatorAddress,
			}
			resp, err := m.keeper.GetStakingQueryServer().UnbondingDelegation(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.Unbond, nil
		},
		UpdateQuery: func(ctx context.Context) any {
			req := &stakingtypes.QueryUnbondingDelegationRequest{
				DelegatorAddr: delegatorAddress,
				ValidatorAddr: validatorAddress,
			}
			resp, err := m.keeper.GetStakingQueryServer().UnbondingDelegation(ctx, req)
			if err != nil {
				return nil
			}
			return resp.Unbond
		},
	}

	connParams := qb.BuildConnectionParams(w, params)
	m.HandleStandardConnection(connParams)
}
