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

// Module implements the staking WebSocket module
type Module struct {
	*common.ModuleHandler
	keeper KeeperInterface
	name   string
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

	delAddr := addrValidator.AccAddress()
	qb := common.NewQueryBuilder(m.keeper, m.Logger)

	params := common.StandardSubscriptionParams{
		Request:        r,
		ValidationFunc: addrValidator.ValidationFunc(),
		SubscriptionKey: common.NewSubscriptionKeyAdapter(
			types.GenerateSubscriptionKey(types.SubscriptionTypeDelegations, delegatorAddress, "", ""),
		),
		InitialQuery: func(ctx context.Context) (any, error) {
			return m.getDelegationResponses(ctx, delAddr), nil
		},
		UpdateQuery: func(ctx context.Context) any {
			return m.getDelegationResponses(ctx, delAddr)
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

	delAddr := delAddrValidator.AccAddress()
	valAddr := valAddrValidator.ValAddress()
	qb := common.NewQueryBuilder(m.keeper, m.Logger)

	params := common.StandardSubscriptionParams{
		Request:        r,
		ValidationFunc: validator.ValidationFunc(),
		SubscriptionKey: common.NewSubscriptionKeyAdapter(
			types.GenerateSubscriptionKey(types.SubscriptionTypeDelegation, delegatorAddress, validatorAddress, ""),
		),
		InitialQuery: func(ctx context.Context) (any, error) {
			return m.getDelegationResponse(ctx, delAddr, valAddr), nil
		},
		UpdateQuery: func(ctx context.Context) any {
			return m.getDelegationResponse(ctx, delAddr, valAddr)
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

	delAddr := addrValidator.AccAddress()
	qb := common.NewQueryBuilder(m.keeper, m.Logger)

	params := common.StandardSubscriptionParams{
		Request:        r,
		ValidationFunc: addrValidator.ValidationFunc(),
		SubscriptionKey: common.NewSubscriptionKeyAdapter(
			types.GenerateSubscriptionKey(types.SubscriptionTypeUnbondingDelegations, delegatorAddress, "", ""),
		),
		InitialQuery: func(ctx context.Context) (any, error) {
			unbondingDelegations, err := m.keeper.GetStakingKeeper().GetAllUnbondingDelegations(ctx, delAddr)
			if err != nil {
				return []stakingtypes.UnbondingDelegation{}, nil
			}
			return unbondingDelegations, nil
		},
		UpdateQuery: func(ctx context.Context) any {
			unbondingDelegations, err := m.keeper.GetStakingKeeper().GetAllUnbondingDelegations(ctx, delAddr)
			if err != nil {
				return []stakingtypes.UnbondingDelegation{}
			}
			return unbondingDelegations
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

	delAddr := delAddrValidator.AccAddress()
	valAddr := valAddrValidator.ValAddress()
	qb := common.NewQueryBuilder(m.keeper, m.Logger)

	params := common.StandardSubscriptionParams{
		Request:        r,
		ValidationFunc: validator.ValidationFunc(),
		SubscriptionKey: common.NewSubscriptionKeyAdapter(
			types.GenerateSubscriptionKey(types.SubscriptionTypeUnbondingDelegation, delegatorAddress, validatorAddress, ""),
		),
		InitialQuery: func(ctx context.Context) (any, error) {
			return m.getUnbondingDelegationResponse(ctx, delAddr, valAddr), nil
		},
		UpdateQuery: func(ctx context.Context) any {
			return m.getUnbondingDelegationResponse(ctx, delAddr, valAddr)
		},
	}

	connParams := qb.BuildConnectionParams(w, params)
	m.HandleStandardConnection(connParams)
}

// getDelegationResponses gets delegation responses for a delegator
func (m *Module) getDelegationResponses(ctx context.Context, delAddr sdk.AccAddress) []stakingtypes.DelegationResponse {
	stakingKeeper := m.keeper.GetStakingKeeper()
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
func (m *Module) getDelegationResponse(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) any {
	stakingKeeper := m.keeper.GetStakingKeeper()
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

// getUnbondingDelegationResponse gets an unbonding delegation response for a specific delegator-validator pair
func (m *Module) getUnbondingDelegationResponse(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) any {
	unbondingDelegation, err := m.keeper.GetStakingKeeper().GetUnbondingDelegation(ctx, delAddr, valAddr)
	if err != nil {
		return common.NotFoundResponse()
	}
	return common.FoundResponse("unbonding_delegation", unbondingDelegation)
}