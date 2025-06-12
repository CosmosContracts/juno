package staking

import (
	"context"
	"net/http"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gorilla/mux"

	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/common"
	"github.com/CosmosContracts/juno/v30/x/stream/types"
)

// UnbondingHandler handles unbonding delegations subscription WebSocket connections
type UnbondingHandler struct {
	common.QueryHandler
	keeper KeeperInterface
}

// NewUnbondingHandler creates a new unbonding handler
func NewUnbondingHandler(keeper KeeperInterface, deps *common.HandlerDependencies) *UnbondingHandler {
	return &UnbondingHandler{
		QueryHandler: common.NewQueryHandler(deps, keeper),
		keeper:       keeper,
	}
}


// HandleUnbondingDelegations handles unbonding delegations subscription WebSocket connections
func (h *UnbondingHandler) HandleUnbondingDelegations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	delegatorAddress := vars["delegator"]

	// Validate address
	addrValidator := common.ValidateAccAddress(delegatorAddress)
	if !addrValidator.IsValid() {
		http.Error(w, addrValidator.Error().Error(), http.StatusBadRequest)
		return
	}

	delAddr := addrValidator.AccAddress()

	params := common.ConnectionParams{
		Writer:         w,
		Request:        r,
		ValidationFunc: addrValidator.ValidationFunc(),
		InitialDataFunc: h.WrapInitialDataFunc(func(ctx context.Context) (any, error) {
			unbondingDelegations, err := h.keeper.GetStakingKeeper().GetAllUnbondingDelegations(ctx, delAddr)
			if err != nil {
				return []stakingtypes.UnbondingDelegation{}, nil
			}
			return unbondingDelegations, nil
		}),
		SubscriptionKey: common.NewSubscriptionKeyAdapter(types.GenerateSubscriptionKey(types.SubscriptionTypeUnbondingDelegations, delegatorAddress, "", "")),
		QueryFunc: h.WrapQueryFunc(func(ctx context.Context) any {
			unbondingDelegations, err := h.keeper.GetStakingKeeper().GetAllUnbondingDelegations(ctx, delAddr)
			if err != nil {
				return []stakingtypes.UnbondingDelegation{}
			}
			return unbondingDelegations
		}),
	}

	h.HandleStandardConnection(params)
}

// HandleUnbondingDelegation handles unbonding delegation subscription WebSocket connections
func (h *UnbondingHandler) HandleUnbondingDelegation(w http.ResponseWriter, r *http.Request) {
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

	delAddr := delAddrValidator.AccAddress()
	valAddr := valAddrValidator.ValAddress()

	params := common.ConnectionParams{
		Writer:         w,
		Request:        r,
		ValidationFunc: validator.ValidationFunc(),
		InitialDataFunc: h.WrapInitialDataFunc(func(ctx context.Context) (any, error) {
			return getUnbondingDelegationResponse(h.keeper, ctx, delAddr, valAddr), nil
		}),
		SubscriptionKey: common.NewSubscriptionKeyAdapter(types.GenerateSubscriptionKey(types.SubscriptionTypeUnbondingDelegation, delegatorAddress, validatorAddress, "")),
		QueryFunc: h.WrapQueryFunc(func(ctx context.Context) any {
			return getUnbondingDelegationResponse(h.keeper, ctx, delAddr, valAddr)
		}),
	}

	h.HandleStandardConnection(params)
}

// getUnbondingDelegationResponse gets an unbonding delegation response for a specific delegator-validator pair
func getUnbondingDelegationResponse(keeper KeeperInterface, ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) any {
	unbondingDelegation, err := keeper.GetStakingKeeper().GetUnbondingDelegation(ctx, delAddr, valAddr)
	if err != nil {
		return common.NotFoundResponse()
	}
	return common.FoundResponse("unbonding_delegation", unbondingDelegation)
}
