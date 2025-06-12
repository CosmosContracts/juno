package bank

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/common"
	"github.com/CosmosContracts/juno/v30/x/stream/types"
)

// AllBalancesHandler handles all balances subscription WebSocket connections
type AllBalancesHandler struct {
	common.QueryHandler
	keeper KeeperInterface
}

// NewAllBalancesHandler creates a new all balances handler
func NewAllBalancesHandler(keeper KeeperInterface, deps *common.HandlerDependencies) *AllBalancesHandler {
	return &AllBalancesHandler{
		QueryHandler: common.NewQueryHandler(deps, keeper),
		keeper:       keeper,
	}
}


// Handle handles all balances subscription WebSocket connections
func (h *AllBalancesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	// Validate address
	addrValidator := common.ValidateAccAddress(address)
	if !addrValidator.IsValid() {
		http.Error(w, addrValidator.Error().Error(), http.StatusBadRequest)
		return
	}

	addr := addrValidator.AccAddress()

	params := common.ConnectionParams{
		Writer:         w,
		Request:        r,
		ValidationFunc: addrValidator.ValidationFunc(),
		InitialDataFunc: h.WrapInitialDataFunc(func(ctx context.Context) (any, error) {
			balances := h.keeper.GetBankKeeper().GetAllBalances(ctx, addr)
			return common.BalancesResponse(balances), nil
		}),
		SubscriptionKey: common.NewSubscriptionKeyAdapter(types.GenerateSubscriptionKey(types.SubscriptionTypeAllBalances, address, "", "")),
		QueryFunc: h.WrapQueryFunc(func(ctx context.Context) any {
			return common.BalancesResponse(h.keeper.GetBankKeeper().GetAllBalances(ctx, addr))
		}),
	}

	h.HandleStandardConnection(params)
}
