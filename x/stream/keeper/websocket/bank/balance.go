package bank

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/common"
	"github.com/CosmosContracts/juno/v30/x/stream/types"
)

// BalanceHandler handles balance subscription WebSocket connections
type BalanceHandler struct {
	common.QueryHandler
	keeper KeeperInterface
}

// NewBalanceHandler creates a new balance handler
func NewBalanceHandler(keeper KeeperInterface, deps *common.HandlerDependencies) *BalanceHandler {
	return &BalanceHandler{
		QueryHandler: common.NewQueryHandler(deps, keeper),
		keeper:       keeper,
	}
}


// Handle handles balance subscription WebSocket connections
func (h *BalanceHandler) Handle(w http.ResponseWriter, r *http.Request) {
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
	if err := h.keeper.ValidateDenom(r.Context(), denom); err != nil {
		http.Error(w, "invalid denom: "+err.Error(), http.StatusBadRequest)
		return
	}

	addr := addrValidator.AccAddress()

	params := common.ConnectionParams{
		Writer:         w,
		Request:        r,
		ValidationFunc: addrValidator.ValidationFunc(),
		InitialDataFunc: h.WrapInitialDataFunc(func(ctx context.Context) (any, error) {
			return h.keeper.GetBankKeeper().GetBalance(ctx, addr, denom), nil
		}),
		SubscriptionKey: common.NewSubscriptionKeyAdapter(types.GenerateSubscriptionKey(types.SubscriptionTypeBalance, address, "", denom)),
		QueryFunc: h.WrapQueryFunc(func(ctx context.Context) any {
			return h.keeper.GetBankKeeper().GetBalance(ctx, addr, denom)
		}),
	}

	h.HandleStandardConnection(params)
}
