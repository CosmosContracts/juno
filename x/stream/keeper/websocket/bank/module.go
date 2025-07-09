package bank

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/common"
	"github.com/CosmosContracts/juno/v30/x/stream/types"
)

// Module implements the bank WebSocket module
type Module struct {
	*common.ModuleHandler
	keeper           KeeperInterface
	name             string
	routeDefinitions []common.RouteDefinition
}

// NewModule creates a new bank module
func NewModule(keeper KeeperInterface, deps *common.HandlerDependencies) *Module {
	m := &Module{
		ModuleHandler: common.NewModuleHandler(deps, keeper),
		keeper:        keeper,
		name:          "bank",
	}

	// Register routes
	m.RegisterRoute("balance", m.handleBalance)
	m.RegisterRoute("all_balances", m.handleAllBalances)
	m.RegisterRoute("spendable_balances", m.handleSpendableBalances)
	m.RegisterRoute("spendable_balance_by_denom", m.handleSpendableBalanceByDenom)
	m.RegisterRoute("total_supply", m.handleTotalSupply)
	m.RegisterRoute("supply_of", m.handleSupplyOf)
	m.RegisterRoute("params", m.handleParams)
	m.RegisterRoute("denoms_metadata", m.handleDenomsMetadata)
	m.RegisterRoute("denom_metadata", m.handleDenomMetadata)
	m.RegisterRoute("denom_owners", m.handleDenomOwners)
	m.RegisterRoute("send_enabled", m.handleSendEnabled)

	// Define route patterns for automatic registration
	m.routeDefinitions = []common.RouteDefinition{
		{Pattern: "/ws/subscribe/bank/balance/{address}/{denom}", Handler: m.handleBalance},
		{Pattern: "/ws/subscribe/bank/balances/{address}", Handler: m.handleAllBalances},
		{Pattern: "/ws/subscribe/bank/spendable-balances/{address}", Handler: m.handleSpendableBalances},
		{Pattern: "/ws/subscribe/bank/spendable-balance/{address}/{denom}", Handler: m.handleSpendableBalanceByDenom},
		{Pattern: "/ws/subscribe/bank/total-supply", Handler: m.handleTotalSupply},
		{Pattern: "/ws/subscribe/bank/supply/{denom}", Handler: m.handleSupplyOf},
		{Pattern: "/ws/subscribe/bank/params", Handler: m.handleParams},
		{Pattern: "/ws/subscribe/bank/denoms-metadata", Handler: m.handleDenomsMetadata},
		{Pattern: "/ws/subscribe/bank/denom-metadata/{denom}", Handler: m.handleDenomMetadata},
		{Pattern: "/ws/subscribe/bank/denom-owners/{denom}", Handler: m.handleDenomOwners},
		{Pattern: "/ws/subscribe/bank/send-enabled", Handler: m.handleSendEnabled},
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

// handleBalance handles balance subscription requests
func (m *Module) handleBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	denom := vars["denom"]

	// Validate address
	addrValidator := common.ValidateAccAddress(address)
	if !addrValidator.IsValid() {
		http.Error(w, addrValidator.Error().Error(), http.StatusBadRequest)
		return
	}

	// Validate denom
	if err := m.keeper.ValidateDenom(r.Context(), denom); err != nil {
		http.Error(w, "invalid denom: "+err.Error(), http.StatusBadRequest)
		return
	}

	qb := common.NewQueryBuilder(m.keeper, m.Logger)

	params := common.StandardSubscriptionParams{
		Request:        r,
		ValidationFunc: addrValidator.ValidationFunc(),
		SubscriptionKey: common.NewSubscriptionKeyAdapter(
			types.GenerateSubscriptionKey(types.SubscriptionTypeBalance, address, "", denom),
		),
		InitialQuery: func(ctx context.Context) (any, error) {
			req := &banktypes.QueryBalanceRequest{
				Address: address,
				Denom:   denom,
			}
			resp, err := m.keeper.GetBankQueryServer().Balance(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.Balance, nil
		},
		UpdateQuery: func(ctx context.Context) any {
			req := &banktypes.QueryBalanceRequest{
				Address: address,
				Denom:   denom,
			}
			resp, err := m.keeper.GetBankQueryServer().Balance(ctx, req)
			if err != nil {
				return nil
			}
			return resp.Balance
		},
	}

	connParams := qb.BuildConnectionParams(w, params)
	m.HandleStandardConnection(connParams)
}

// handleAllBalances handles all balances subscription requests
func (m *Module) handleAllBalances(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	// Validate address
	addrValidator := common.ValidateAccAddress(address)
	if !addrValidator.IsValid() {
		http.Error(w, addrValidator.Error().Error(), http.StatusBadRequest)
		return
	}

	qb := common.NewQueryBuilder(m.keeper, m.Logger)

	params := common.StandardSubscriptionParams{
		Request:        r,
		ValidationFunc: addrValidator.ValidationFunc(),
		SubscriptionKey: common.NewSubscriptionKeyAdapter(
			types.GenerateSubscriptionKey(types.SubscriptionTypeAllBalances, address, "", ""),
		),
		InitialQuery: func(ctx context.Context) (any, error) {
			req := &banktypes.QueryAllBalancesRequest{
				Address: address,
				Pagination: &query.PageRequest{
					Limit: 1000, // Get all balances up to 1000
				},
			}
			resp, err := m.keeper.GetBankQueryServer().AllBalances(ctx, req)
			if err != nil {
				return nil, err
			}
			return common.BalancesResponse(resp.Balances), nil
		},
		UpdateQuery: func(ctx context.Context) any {
			req := &banktypes.QueryAllBalancesRequest{
				Address: address,
				Pagination: &query.PageRequest{
					Limit: 1000,
				},
			}
			resp, err := m.keeper.GetBankQueryServer().AllBalances(ctx, req)
			if err != nil {
				return nil
			}
			return common.BalancesResponse(resp.Balances)
		},
	}

	connParams := qb.BuildConnectionParams(w, params)
	m.HandleStandardConnection(connParams)
}

// handleSpendableBalances handles spendable balances subscription requests
func (m *Module) handleSpendableBalances(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	// Validate address
	addrValidator := common.ValidateAccAddress(address)
	if !addrValidator.IsValid() {
		http.Error(w, addrValidator.Error().Error(), http.StatusBadRequest)
		return
	}

	qb := common.NewQueryBuilder(m.keeper, m.Logger)

	params := common.StandardSubscriptionParams{
		Request:        r,
		ValidationFunc: addrValidator.ValidationFunc(),
		SubscriptionKey: common.NewSubscriptionKeyAdapter(
			types.GenerateSubscriptionKey(types.SubscriptionTypeSpendableBalances, address, "", ""),
		),
		InitialQuery: func(ctx context.Context) (any, error) {
			req := &banktypes.QuerySpendableBalancesRequest{
				Address: address,
				Pagination: &query.PageRequest{
					Limit: 1000,
				},
			}
			resp, err := m.keeper.GetBankQueryServer().SpendableBalances(ctx, req)
			if err != nil {
				return nil, err
			}
			return common.BalancesResponse(resp.Balances), nil
		},
		UpdateQuery: func(ctx context.Context) any {
			req := &banktypes.QuerySpendableBalancesRequest{
				Address: address,
				Pagination: &query.PageRequest{
					Limit: 1000,
				},
			}
			resp, err := m.keeper.GetBankQueryServer().SpendableBalances(ctx, req)
			if err != nil {
				return nil
			}
			return common.BalancesResponse(resp.Balances)
		},
	}

	connParams := qb.BuildConnectionParams(w, params)
	m.HandleStandardConnection(connParams)
}

// handleSpendableBalanceByDenom handles spendable balance by denom subscription requests
func (m *Module) handleSpendableBalanceByDenom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	denom := vars["denom"]

	// Validate address
	addrValidator := common.ValidateAccAddress(address)
	if !addrValidator.IsValid() {
		http.Error(w, addrValidator.Error().Error(), http.StatusBadRequest)
		return
	}

	// Validate denom
	if err := m.keeper.ValidateDenom(r.Context(), denom); err != nil {
		http.Error(w, "invalid denom: "+err.Error(), http.StatusBadRequest)
		return
	}

	qb := common.NewQueryBuilder(m.keeper, m.Logger)

	params := common.StandardSubscriptionParams{
		Request:        r,
		ValidationFunc: addrValidator.ValidationFunc(),
		SubscriptionKey: common.NewSubscriptionKeyAdapter(
			types.GenerateSubscriptionKey(types.SubscriptionTypeSpendableBalanceByDenom, address, "", denom),
		),
		InitialQuery: func(ctx context.Context) (any, error) {
			req := &banktypes.QuerySpendableBalanceByDenomRequest{
				Address: address,
				Denom:   denom,
			}
			resp, err := m.keeper.GetBankQueryServer().SpendableBalanceByDenom(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.Balance, nil
		},
		UpdateQuery: func(ctx context.Context) any {
			req := &banktypes.QuerySpendableBalanceByDenomRequest{
				Address: address,
				Denom:   denom,
			}
			resp, err := m.keeper.GetBankQueryServer().SpendableBalanceByDenom(ctx, req)
			if err != nil {
				return nil
			}
			return resp.Balance
		},
	}

	connParams := qb.BuildConnectionParams(w, params)
	m.HandleStandardConnection(connParams)
}

// handleTotalSupply handles total supply subscription requests
func (m *Module) handleTotalSupply(w http.ResponseWriter, r *http.Request) {
	qb := common.NewQueryBuilder(m.keeper, m.Logger)

	params := common.StandardSubscriptionParams{
		Request: r,
		ValidationFunc: func() error {
			return nil // No validation needed
		},
		SubscriptionKey: common.NewSubscriptionKeyAdapter(
			types.GenerateSubscriptionKey(types.SubscriptionTypeTotalSupply, "", "", ""),
		),
		InitialQuery: func(ctx context.Context) (any, error) {
			req := &banktypes.QueryTotalSupplyRequest{
				Pagination: &query.PageRequest{
					Limit: 1000,
				},
			}
			resp, err := m.keeper.GetBankQueryServer().TotalSupply(ctx, req)
			if err != nil {
				return nil, err
			}
			return common.BalancesResponse(resp.Supply), nil
		},
		UpdateQuery: func(ctx context.Context) any {
			req := &banktypes.QueryTotalSupplyRequest{
				Pagination: &query.PageRequest{
					Limit: 1000,
				},
			}
			resp, err := m.keeper.GetBankQueryServer().TotalSupply(ctx, req)
			if err != nil {
				return nil
			}
			return common.BalancesResponse(resp.Supply)
		},
	}

	connParams := qb.BuildConnectionParams(w, params)
	m.HandleStandardConnection(connParams)
}

// handleSupplyOf handles supply of a specific denom subscription requests
func (m *Module) handleSupplyOf(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	denom := vars["denom"]

	// Validate denom
	if err := m.keeper.ValidateDenom(r.Context(), denom); err != nil {
		http.Error(w, "invalid denom: "+err.Error(), http.StatusBadRequest)
		return
	}

	qb := common.NewQueryBuilder(m.keeper, m.Logger)

	params := common.StandardSubscriptionParams{
		Request: r,
		ValidationFunc: func() error {
			return m.keeper.ValidateDenom(r.Context(), denom)
		},
		SubscriptionKey: common.NewSubscriptionKeyAdapter(
			types.GenerateSubscriptionKey(types.SubscriptionTypeSupplyOf, "", "", denom),
		),
		InitialQuery: func(ctx context.Context) (any, error) {
			req := &banktypes.QuerySupplyOfRequest{
				Denom: denom,
			}
			resp, err := m.keeper.GetBankQueryServer().SupplyOf(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.Amount, nil
		},
		UpdateQuery: func(ctx context.Context) any {
			req := &banktypes.QuerySupplyOfRequest{
				Denom: denom,
			}
			resp, err := m.keeper.GetBankQueryServer().SupplyOf(ctx, req)
			if err != nil {
				return nil
			}
			return resp.Amount
		},
	}

	connParams := qb.BuildConnectionParams(w, params)
	m.HandleStandardConnection(connParams)
}

// handleParams handles bank params subscription requests
func (m *Module) handleParams(w http.ResponseWriter, r *http.Request) {
	qb := common.NewQueryBuilder(m.keeper, m.Logger)

	params := common.StandardSubscriptionParams{
		Request: r,
		ValidationFunc: func() error {
			return nil // No validation needed
		},
		SubscriptionKey: common.NewSubscriptionKeyAdapter(
			types.GenerateSubscriptionKey(types.SubscriptionTypeParams, "", "", ""),
		),
		InitialQuery: func(ctx context.Context) (any, error) {
			req := &banktypes.QueryParamsRequest{}
			resp, err := m.keeper.GetBankQueryServer().Params(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.Params, nil
		},
		UpdateQuery: func(ctx context.Context) any {
			req := &banktypes.QueryParamsRequest{}
			resp, err := m.keeper.GetBankQueryServer().Params(ctx, req)
			if err != nil {
				return nil
			}
			return resp.Params
		},
	}

	connParams := qb.BuildConnectionParams(w, params)
	m.HandleStandardConnection(connParams)
}

// handleDenomsMetadata handles all denoms metadata subscription requests
func (m *Module) handleDenomsMetadata(w http.ResponseWriter, r *http.Request) {
	qb := common.NewQueryBuilder(m.keeper, m.Logger)

	params := common.StandardSubscriptionParams{
		Request: r,
		ValidationFunc: func() error {
			return nil // No validation needed
		},
		SubscriptionKey: common.NewSubscriptionKeyAdapter(
			types.GenerateSubscriptionKey(types.SubscriptionTypeDenomsMetadata, "", "", ""),
		),
		InitialQuery: func(ctx context.Context) (any, error) {
			req := &banktypes.QueryDenomsMetadataRequest{
				Pagination: &query.PageRequest{
					Limit: 1000,
				},
			}
			resp, err := m.keeper.GetBankQueryServer().DenomsMetadata(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.Metadatas, nil
		},
		UpdateQuery: func(ctx context.Context) any {
			req := &banktypes.QueryDenomsMetadataRequest{
				Pagination: &query.PageRequest{
					Limit: 1000,
				},
			}
			resp, err := m.keeper.GetBankQueryServer().DenomsMetadata(ctx, req)
			if err != nil {
				return nil
			}
			return resp.Metadatas
		},
	}

	connParams := qb.BuildConnectionParams(w, params)
	m.HandleStandardConnection(connParams)
}

// handleDenomMetadata handles denom metadata subscription requests
func (m *Module) handleDenomMetadata(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	denom := vars["denom"]

	// Validate denom
	if err := m.keeper.ValidateDenom(r.Context(), denom); err != nil {
		http.Error(w, "invalid denom: "+err.Error(), http.StatusBadRequest)
		return
	}

	qb := common.NewQueryBuilder(m.keeper, m.Logger)

	params := common.StandardSubscriptionParams{
		Request: r,
		ValidationFunc: func() error {
			return m.keeper.ValidateDenom(r.Context(), denom)
		},
		SubscriptionKey: common.NewSubscriptionKeyAdapter(
			types.GenerateSubscriptionKey(types.SubscriptionTypeDenomMetadata, "", "", denom),
		),
		InitialQuery: func(ctx context.Context) (any, error) {
			req := &banktypes.QueryDenomMetadataRequest{
				Denom: denom,
			}
			resp, err := m.keeper.GetBankQueryServer().DenomMetadata(ctx, req)
			if err != nil {
				if status.Code(err) == codes.NotFound {
					return nil, fmt.Errorf("metadata not found for denom %s", denom)
				}
				return nil, err
			}
			return resp.Metadata, nil
		},
		UpdateQuery: func(ctx context.Context) any {
			req := &banktypes.QueryDenomMetadataRequest{
				Denom: denom,
			}
			resp, err := m.keeper.GetBankQueryServer().DenomMetadata(ctx, req)
			if err != nil {
				return nil
			}
			return resp.Metadata
		},
	}

	connParams := qb.BuildConnectionParams(w, params)
	m.HandleStandardConnection(connParams)
}

// handleDenomOwners handles denom owners subscription requests
func (m *Module) handleDenomOwners(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	denom := vars["denom"]

	// Validate denom
	if err := m.keeper.ValidateDenom(r.Context(), denom); err != nil {
		http.Error(w, "invalid denom: "+err.Error(), http.StatusBadRequest)
		return
	}

	qb := common.NewQueryBuilder(m.keeper, m.Logger)

	params := common.StandardSubscriptionParams{
		Request: r,
		ValidationFunc: func() error {
			return m.keeper.ValidateDenom(r.Context(), denom)
		},
		SubscriptionKey: common.NewSubscriptionKeyAdapter(
			types.GenerateSubscriptionKey(types.SubscriptionTypeDenomOwners, "", "", denom),
		),
		InitialQuery: func(ctx context.Context) (any, error) {
			req := &banktypes.QueryDenomOwnersRequest{
				Denom: denom,
				Pagination: &query.PageRequest{
					Limit: 1000,
				},
			}
			resp, err := m.keeper.GetBankQueryServer().DenomOwners(ctx, req)
			if err != nil {
				return nil, err
			}

			return resp.DenomOwners, nil
		},
		UpdateQuery: func(ctx context.Context) any {
			req := &banktypes.QueryDenomOwnersRequest{
				Denom: denom,
				Pagination: &query.PageRequest{
					Limit: 1000,
				},
			}
			resp, err := m.keeper.GetBankQueryServer().DenomOwners(ctx, req)
			if err != nil {
				return nil
			}
			return resp
		},
	}

	connParams := qb.BuildConnectionParams(w, params)
	m.HandleStandardConnection(connParams)
}

// handleSendEnabled handles send enabled subscription requests
func (m *Module) handleSendEnabled(w http.ResponseWriter, r *http.Request) {
	qb := common.NewQueryBuilder(m.keeper, m.Logger)

	params := common.StandardSubscriptionParams{
		Request: r,
		ValidationFunc: func() error {
			return nil // No validation needed
		},
		SubscriptionKey: common.NewSubscriptionKeyAdapter(
			types.GenerateSubscriptionKey(types.SubscriptionTypeSendEnabled, "", "", ""),
		),
		InitialQuery: func(ctx context.Context) (any, error) {
			req := &banktypes.QuerySendEnabledRequest{
				Denoms: []string{}, // Get all
				Pagination: &query.PageRequest{
					Limit: 1000,
				},
			}
			resp, err := m.keeper.GetBankQueryServer().SendEnabled(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.SendEnabled, nil
		},
		UpdateQuery: func(ctx context.Context) any {
			req := &banktypes.QuerySendEnabledRequest{
				Denoms: []string{},
				Pagination: &query.PageRequest{
					Limit: 1000,
				},
			}
			resp, err := m.keeper.GetBankQueryServer().SendEnabled(ctx, req)
			if err != nil {
				return nil
			}
			return resp.SendEnabled
		},
	}

	connParams := qb.BuildConnectionParams(w, params)
	m.HandleStandardConnection(connParams)
}
