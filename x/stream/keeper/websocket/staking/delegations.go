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

// DelegationsHandler handles delegations subscription WebSocket connections
type DelegationsHandler struct {
	common.QueryHandler
	keeper KeeperInterface
}

// NewDelegationsHandler creates a new delegations handler
func NewDelegationsHandler(keeper KeeperInterface, deps *common.HandlerDependencies) *DelegationsHandler {
	return &DelegationsHandler{
		QueryHandler: common.NewQueryHandler(deps, keeper),
		keeper:       keeper,
	}
}


// Handle handles delegations subscription WebSocket connections
func (h *DelegationsHandler) Handle(w http.ResponseWriter, r *http.Request) {
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
			return getDelegationResponses(h.keeper, ctx, delAddr), nil
		}),
		SubscriptionKey: common.NewSubscriptionKeyAdapter(types.GenerateSubscriptionKey(types.SubscriptionTypeDelegations, delegatorAddress, "", "")),
		QueryFunc: h.WrapQueryFunc(func(ctx context.Context) any {
			return getDelegationResponses(h.keeper, ctx, delAddr)
		}),
	}

	h.HandleStandardConnection(params)
}

// HandleDelegation handles delegation subscription WebSocket connections
func (h *DelegationsHandler) HandleDelegation(w http.ResponseWriter, r *http.Request) {
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
			return getDelegationResponse(h.keeper, ctx, delAddr, valAddr), nil
		}),
		SubscriptionKey: common.NewSubscriptionKeyAdapter(types.GenerateSubscriptionKey(types.SubscriptionTypeDelegation, delegatorAddress, validatorAddress, "")),
		QueryFunc: h.WrapQueryFunc(func(ctx context.Context) any {
			return getDelegationResponse(h.keeper, ctx, delAddr, valAddr)
		}),
	}

	h.HandleStandardConnection(params)
}

// getDelegationResponses gets delegation responses for a delegator
func getDelegationResponses(keeper KeeperInterface, ctx context.Context, delAddr sdk.AccAddress) []stakingtypes.DelegationResponse {
	stakingKeeper := keeper.GetStakingKeeper()
	delegations, err := stakingKeeper.GetAllDelegatorDelegations(ctx, delAddr)
	if err != nil {
		return []stakingtypes.DelegationResponse{}
	}

	bondDenom, err := stakingKeeper.BondDenom(ctx)
	if err != nil {
		return []stakingtypes.DelegationResponse{}
	}

	var delegationResponses []stakingtypes.DelegationResponse
	for _, delegation := range delegations {
		valAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
		if err != nil {
			continue
		}

		validator, err := stakingKeeper.GetValidator(ctx, valAddr)
		if err == nil {
			delegationResponses = append(delegationResponses, stakingtypes.DelegationResponse{
				Delegation: delegation,
				Balance:    sdk.NewCoin(bondDenom, validator.TokensFromShares(delegation.Shares).TruncateInt()),
			})
		}
	}
	return delegationResponses
}

// getDelegationResponse gets a delegation response for a specific delegator-validator pair
func getDelegationResponse(keeper KeeperInterface, ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) any {
	stakingKeeper := keeper.GetStakingKeeper()
	delegation, err := stakingKeeper.GetDelegation(ctx, delAddr, valAddr)

	if err != nil {
		return common.NotFoundResponse()
	}

	bondDenom, bondErr := stakingKeeper.BondDenom(ctx)
	if bondErr != nil {
		return common.NotFoundResponse()
	}

	validator, valErr := stakingKeeper.GetValidator(ctx, valAddr)
	if valErr != nil {
		return common.NotFoundResponse()
	}

	delegationResponse := stakingtypes.DelegationResponse{
		Delegation: delegation,
		Balance:    sdk.NewCoin(bondDenom, validator.TokensFromShares(delegation.Shares).TruncateInt()),
	}
	return common.FoundResponse("delegation", delegationResponse)
}
