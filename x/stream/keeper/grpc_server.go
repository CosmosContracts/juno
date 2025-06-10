package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/CosmosContracts/juno/v30/x/stream/types"
)

var _ types.QueryServer = queryServer{}

func NewQueryServerImpl(k Keeper) types.QueryServer {
	return queryServer{k}
}

type queryServer struct {
	k Keeper
}

// StreamBalance implements the streaming balance gRPC endpoint
func (q queryServer) StreamBalance(req *types.StreamBalanceRequest, stream types.Query_StreamBalanceServer) error {
	streamCtx := stream.Context()
	ctx, err := q.k.GetQueryContext()
	if err != nil {
		return err
	}

	// Validate request
	if req.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if req.Denom == "" {
		return fmt.Errorf("denom cannot be empty")
	}

	// Validate address format
	if _, err := sdk.AccAddressFromBech32(req.Address); err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}

	// Send initial response with current balance
	// The context here will be injected by our interceptor with proper SDK values
	balance := q.k.bankKeeper.GetBalance(ctx, sdk.MustAccAddressFromBech32(req.Address), req.Denom)
	if err := stream.Send(&types.StreamBalanceResponse{Balance: &balance}); err != nil {
		return err
	}

	// Create subscription
	subKey := types.GenerateSubscriptionKey(types.SubscriptionTypeBalance, req.Address, "", req.Denom)
	sendCh := make(chan any, 32)
	subscriber := q.k.registry.Subscribe(subKey, streamCtx, sendCh)
	defer q.k.registry.Unsubscribe(subscriber)

	// Stream updates
	for {
		select {
		case <-streamCtx.Done():
			return streamCtx.Err()
		case update := <-sendCh:
			if _, ok := update.(types.StreamEvent); ok {
				// Re-query the current balance using the latest query context
				queryCtx, err := q.k.GetQueryContext()
				if err != nil {
					return err
				}
				balance := q.k.bankKeeper.GetBalance(queryCtx, sdk.MustAccAddressFromBech32(req.Address), req.Denom)
				if err := stream.Send(&types.StreamBalanceResponse{Balance: &balance}); err != nil {
					return err
				}
			}
		}
	}
}

// StreamAllBalances implements the streaming all balances gRPC endpoint
func (q queryServer) StreamAllBalances(req *types.StreamAllBalancesRequest, stream types.Query_StreamAllBalancesServer) error {
	streamCtx := stream.Context()
	ctx, err := q.k.GetQueryContext()
	if err != nil {
		return err
	}

	// Validate request
	if req.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}

	// Validate address format
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}

	// Send initial response with all balances
	balances := q.k.bankKeeper.GetAllBalances(ctx, addr)
	balancePointers := make([]*sdk.Coin, len(balances))
	for i := range balances {
		balancePointers[i] = &balances[i]
	}
	if err := stream.Send(&types.StreamAllBalancesResponse{Balances: balancePointers}); err != nil {
		return err
	}

	// Create subscription
	subKey := types.GenerateSubscriptionKey(types.SubscriptionTypeAllBalances, req.Address, "", "")
	sendCh := make(chan any, 32)
	subscriber := q.k.registry.Subscribe(subKey, streamCtx, sendCh)
	defer q.k.registry.Unsubscribe(subscriber)

	// Stream updates
	for {
		select {
		case <-streamCtx.Done():
			return streamCtx.Err()
		case update := <-sendCh:
			if _, ok := update.(types.StreamEvent); ok {
				queryCtx, err := q.k.GetQueryContext()
				if err != nil {
					return err
				}
				balances := q.k.bankKeeper.GetAllBalances(queryCtx, sdk.MustAccAddressFromBech32(req.Address))
				balancePointers := make([]*sdk.Coin, len(balances))
				for i := range balances {
					balancePointers[i] = &balances[i]
				}
				if err := stream.Send(&types.StreamAllBalancesResponse{Balances: balancePointers}); err != nil {
					return err
				}
			}
		}
	}
}

// StreamDelegations implements the streaming delegations gRPC endpoint
func (q queryServer) StreamDelegations(req *types.StreamDelegationsRequest, stream types.Query_StreamDelegationsServer) error {
	streamCtx := stream.Context()
	ctx, err := q.k.GetQueryContext()
	if err != nil {
		return err
	}

	// Validate request
	if req.DelegatorAddress == "" {
		return fmt.Errorf("delegator address cannot be empty")
	}

	// Validate address format
	delAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddress)
	if err != nil {
		return fmt.Errorf("invalid delegator address: %w", err)
	}

	// Get initial delegations
	delegations, err := q.k.stakingKeeper.GetAllDelegatorDelegations(ctx, delAddr)
	if err != nil {
		return fmt.Errorf("failed to get delegations: %w", err)
	}

	// Convert to delegation responses
	bondDenom, err := q.k.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return fmt.Errorf("failed to get bond denom: %w", err)
	}

	var delegationResponses []*stakingtypes.DelegationResponse
	for _, delegation := range delegations {
		valAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
		if err != nil {
			continue
		}

		validator, err := q.k.stakingKeeper.GetValidator(ctx, valAddr)
		if err == nil {
			delegationResponses = append(delegationResponses, &stakingtypes.DelegationResponse{
				Delegation: delegation,
				Balance:    sdk.NewCoin(bondDenom, validator.TokensFromShares(delegation.Shares).TruncateInt()),
			})
		}
	}

	if err := stream.Send(&types.StreamDelegationsResponse{Delegations: delegationResponses}); err != nil {
		return err
	}

	// Create subscription
	subKey := types.GenerateSubscriptionKey(types.SubscriptionTypeDelegations, req.DelegatorAddress, "", "")
	sendCh := make(chan any, 32)
	subscriber := q.k.registry.Subscribe(subKey, streamCtx, sendCh)
	defer q.k.registry.Unsubscribe(subscriber)

	// Stream updates
	for {
		select {
		case <-streamCtx.Done():
			return streamCtx.Err()
		case update := <-sendCh:
			if _, ok := update.(types.StreamEvent); ok {
				queryCtx, err := q.k.GetQueryContext()
				if err != nil {
					return err
				}
				// Re-query delegations
				delegations, err := q.k.stakingKeeper.GetAllDelegatorDelegations(queryCtx, delAddr)
				if err != nil {
					continue
				}

				// Convert to delegation responses
				var delegationResponses []*stakingtypes.DelegationResponse
				for _, delegation := range delegations {
					valAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
					if err != nil {
						continue
					}

					validator, err := q.k.stakingKeeper.GetValidator(queryCtx, valAddr)
					if err == nil {
						delegationResponses = append(delegationResponses, &stakingtypes.DelegationResponse{
							Delegation: delegation,
							Balance:    sdk.NewCoin(bondDenom, validator.TokensFromShares(delegation.Shares).TruncateInt()),
						})
					}
				}

				if err := stream.Send(&types.StreamDelegationsResponse{Delegations: delegationResponses}); err != nil {
					return err
				}
			}
		}
	}
}

// StreamDelegation implements the streaming single delegation gRPC endpoint
func (q queryServer) StreamDelegation(req *types.StreamDelegationRequest, stream types.Query_StreamDelegationServer) error {
	streamCtx := stream.Context()
	ctx, err := q.k.GetQueryContext()
	if err != nil {
		return err
	}

	// Validate request
	if req.DelegatorAddress == "" {
		return fmt.Errorf("delegator address cannot be empty")
	}
	if req.ValidatorAddress == "" {
		return fmt.Errorf("validator address cannot be empty")
	}

	// Validate addresses
	delAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddress)
	if err != nil {
		return fmt.Errorf("invalid delegator address: %w", err)
	}
	valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		return fmt.Errorf("invalid validator address: %w", err)
	}

	// Get initial delegation
	delegation, err := q.k.stakingKeeper.GetDelegation(ctx, delAddr, valAddr)
	if err != nil {
		return fmt.Errorf("delegation not found: %w", err)
	}

	bondDenom, err := q.k.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return fmt.Errorf("failed to get bond denom: %w", err)
	}

	validator, err := q.k.stakingKeeper.GetValidator(ctx, valAddr)
	if err != nil {
		return fmt.Errorf("validator not found: %w", err)
	}

	response := &types.StreamDelegationResponse{
		Delegation: &stakingtypes.DelegationResponse{
			Delegation: delegation,
			Balance:    sdk.NewCoin(bondDenom, validator.TokensFromShares(delegation.Shares).TruncateInt()),
		},
	}

	if err := stream.Send(response); err != nil {
		return err
	}

	// Create subscription
	subKey := types.GenerateSubscriptionKey(types.SubscriptionTypeDelegation, req.DelegatorAddress, req.ValidatorAddress, "")
	sendCh := make(chan any, 32)
	subscriber := q.k.registry.Subscribe(subKey, streamCtx, sendCh)
	defer q.k.registry.Unsubscribe(subscriber)

	// Stream updates
	for {
		select {
		case <-streamCtx.Done():
			return streamCtx.Err()
		case update := <-sendCh:
			if _, ok := update.(types.StreamEvent); ok {
				queryCtx, err := q.k.GetQueryContext()
				if err != nil {
					return err
				}
				// Re-query delegation
				delegation, err := q.k.stakingKeeper.GetDelegation(queryCtx, delAddr, valAddr)
				if err != nil {
					continue
				}

				validator, err := q.k.stakingKeeper.GetValidator(queryCtx, valAddr)
				if err != nil {
					continue
				}

				response := &types.StreamDelegationResponse{
					Delegation: &stakingtypes.DelegationResponse{
						Delegation: delegation,
						Balance:    sdk.NewCoin(bondDenom, validator.TokensFromShares(delegation.Shares).TruncateInt()),
					},
				}

				if err := stream.Send(response); err != nil {
					return err
				}
			}
		}
	}
}

// StreamUnbondingDelegations implements the streaming unbonding delegations gRPC endpoint
func (q queryServer) StreamUnbondingDelegations(req *types.StreamUnbondingDelegationsRequest, stream types.Query_StreamUnbondingDelegationsServer) error {
	streamCtx := stream.Context()
	ctx, err := q.k.GetQueryContext()
	if err != nil {
		return err
	}

	// Validate request
	if req.DelegatorAddress == "" {
		return fmt.Errorf("delegator address cannot be empty")
	}

	// Validate address format
	delAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddress)
	if err != nil {
		return fmt.Errorf("invalid delegator address: %w", err)
	}

	// Get initial unbonding delegations
	unbondingDelegations, err := q.k.stakingKeeper.GetAllUnbondingDelegations(ctx, delAddr)
	if err != nil {
		return fmt.Errorf("failed to get unbonding delegations: %w", err)
	}

	// Convert to pointer slice
	unbondingDelegationPointers := make([]*stakingtypes.UnbondingDelegation, len(unbondingDelegations))
	for i := range unbondingDelegations {
		unbondingDelegationPointers[i] = &unbondingDelegations[i]
	}

	if err := stream.Send(&types.StreamUnbondingDelegationsResponse{Delegations: unbondingDelegationPointers}); err != nil {
		return err
	}

	// Create subscription
	subKey := types.GenerateSubscriptionKey(types.SubscriptionTypeUnbondingDelegations, req.DelegatorAddress, "", "")
	sendCh := make(chan any, 32)
	subscriber := q.k.registry.Subscribe(subKey, streamCtx, sendCh)
	defer q.k.registry.Unsubscribe(subscriber)

	// Stream updates
	for {
		select {
		case <-streamCtx.Done():
			return streamCtx.Err()
		case update := <-sendCh:
			if _, ok := update.(types.StreamEvent); ok {
				queryCtx, err := q.k.GetQueryContext()
				if err != nil {
					return err
				}
				// Re-query unbonding delegations
				unbondingDelegations, err := q.k.stakingKeeper.GetAllUnbondingDelegations(queryCtx, delAddr)
				if err != nil {
					continue
				}

				// Convert to pointer slice
				unbondingDelegationPointers := make([]*stakingtypes.UnbondingDelegation, len(unbondingDelegations))
				for i := range unbondingDelegations {
					unbondingDelegationPointers[i] = &unbondingDelegations[i]
				}

				if err := stream.Send(&types.StreamUnbondingDelegationsResponse{Delegations: unbondingDelegationPointers}); err != nil {
					return err
				}
			}
		}
	}
}

// StreamUnbondingDelegation implements the streaming single unbonding delegation gRPC endpoint
func (q queryServer) StreamUnbondingDelegation(req *types.StreamUnbondingDelegationRequest, stream types.Query_StreamUnbondingDelegationServer) error {
	streamCtx := stream.Context()
	ctx, err := q.k.GetQueryContext()
	if err != nil {
		return err
	}

	// Validate request
	if req.DelegatorAddress == "" {
		return fmt.Errorf("delegator address cannot be empty")
	}
	if req.ValidatorAddress == "" {
		return fmt.Errorf("validator address cannot be empty")
	}

	// Validate addresses
	delAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddress)
	if err != nil {
		return fmt.Errorf("invalid delegator address: %w", err)
	}
	valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		return fmt.Errorf("invalid validator address: %w", err)
	}

	// Get initial unbonding delegation
	unbondingDelegation, err := q.k.stakingKeeper.GetUnbondingDelegation(ctx, delAddr, valAddr)
	if err != nil {
		return fmt.Errorf("unbonding delegation not found: %w", err)
	}

	response := &types.StreamUnbondingDelegationResponse{
		Delegation: &unbondingDelegation,
	}

	if err := stream.Send(response); err != nil {
		return err
	}

	// Create subscription
	subKey := types.GenerateSubscriptionKey(types.SubscriptionTypeUnbondingDelegation, req.DelegatorAddress, req.ValidatorAddress, "")
	sendCh := make(chan any, 32)
	subscriber := q.k.registry.Subscribe(subKey, streamCtx, sendCh)
	defer q.k.registry.Unsubscribe(subscriber)

	// Stream updates
	for {
		select {
		case <-streamCtx.Done():
			return streamCtx.Err()
		case update := <-sendCh:
			if _, ok := update.(types.StreamEvent); ok {
				queryCtx, err := q.k.GetQueryContext()
				if err != nil {
					return err
				}
				// Re-query unbonding delegation
				unbondingDelegation, err := q.k.stakingKeeper.GetUnbondingDelegation(queryCtx, delAddr, valAddr)
				if err != nil {
					continue
				}

				response := &types.StreamUnbondingDelegationResponse{
					Delegation: &unbondingDelegation,
				}

				if err := stream.Send(response); err != nil {
					return err
				}
			}
		}
	}
}
