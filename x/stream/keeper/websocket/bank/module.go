package bank

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/common"
	"github.com/CosmosContracts/juno/v30/x/stream/types"
)

// Module implements the bank WebSocket module
type Module struct {
	*common.ModuleHandler
	keeper KeeperInterface
	name   string
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

	addr := addrValidator.AccAddress()
	qb := common.NewQueryBuilder(m.keeper, m.Logger)

	params := common.StandardSubscriptionParams{
		Request:        r,
		ValidationFunc: addrValidator.ValidationFunc(),
		SubscriptionKey: common.NewSubscriptionKeyAdapter(
			types.GenerateSubscriptionKey(types.SubscriptionTypeBalance, address, "", denom),
		),
		InitialQuery: func(ctx context.Context) (any, error) {
			return m.keeper.GetBankKeeper().GetBalance(ctx, addr, denom), nil
		},
		UpdateQuery: func(ctx context.Context) any {
			return m.keeper.GetBankKeeper().GetBalance(ctx, addr, denom)
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

	addr := addrValidator.AccAddress()
	qb := common.NewQueryBuilder(m.keeper, m.Logger)

	params := common.StandardSubscriptionParams{
		Request:        r,
		ValidationFunc: addrValidator.ValidationFunc(),
		SubscriptionKey: common.NewSubscriptionKeyAdapter(
			types.GenerateSubscriptionKey(types.SubscriptionTypeAllBalances, address, "", ""),
		),
		InitialQuery: func(ctx context.Context) (any, error) {
			balances := m.keeper.GetBankKeeper().GetAllBalances(ctx, addr)
			return common.BalancesResponse(balances), nil
		},
		UpdateQuery: func(ctx context.Context) any {
			return common.BalancesResponse(m.keeper.GetBankKeeper().GetAllBalances(ctx, addr))
		},
	}

	connParams := qb.BuildConnectionParams(w, params)
	m.HandleStandardConnection(connParams)
}