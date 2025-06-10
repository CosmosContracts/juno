package types

import (
	"context"
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	abci "github.com/cometbft/cometbft/abci/types"
)

// StreamingListener implements the ABCIListener interface for the stream module
type StreamingListener struct {
	logger log.Logger
	intake chan<- StreamEvent
}

// NewStreamingListener creates a new streaming listener
func NewStreamingListener(intake chan<- StreamEvent) *StreamingListener {
	return &StreamingListener{
		logger: log.NewNopLogger(),
		intake: intake,
	}
}

// WithLogger sets the logger for the streaming listener
func (l *StreamingListener) WithLogger(logger log.Logger) *StreamingListener {
	l.logger = logger
	return l
}

// ListenFinalizeBlock implements the ABCIListener interface
func (l *StreamingListener) ListenFinalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	return nil
}

// ListenCommit implements the ABCIListener interface
func (l *StreamingListener) ListenCommit(ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
	return nil
}

// OnWrite implements the ABCIListener interface
func (l *StreamingListener) OnWrite(storeKey storetypes.StoreKey, key []byte, value []byte, delete bool) error {
	// Parse the key and emit events based on the store and key prefix
	event, err := l.parseStoreEvent(storeKey.Name(), key, value, delete)
	if err != nil {
		l.logger.Error("failed to parse store event", "error", err, "store", storeKey.Name())
		return nil
	}

	if event == nil {
		return nil // Not a key we care about
	}

	// Non-blocking send to intake channel
	select {
	case l.intake <- *event:
	default:
		l.logger.Warn("intake channel full, dropping event", "event", event)
	}

	return nil
}

// parseStoreEvent parses a store event and returns a StreamEvent if it matches our criteria
func (l *StreamingListener) parseStoreEvent(storeName string, key []byte, value []byte, delete bool) (*StreamEvent, error) {
	switch storeName {
	case "bank":
		return l.parseBankEvent(key, value, delete)
	case "staking":
		return l.parseStakingEvent(key, value, delete)
	default:
		return nil, nil // Not a store we care about
	}
}

// parseBankEvent parses bank module events
func (l *StreamingListener) parseBankEvent(key []byte, value []byte, delete bool) (*StreamEvent, error) {
	if len(key) == 0 {
		return nil, nil
	}

	// Check if this is a balance key (prefix 0x02)
	if key[0] != BankBalancesPrefix[0] {
		return nil, nil
	}

	// Parse bank balance key format: prefix + address_length + address + denom
	if len(key) < 2 {
		return nil, fmt.Errorf("invalid bank balance key length")
	}

	addrLen := key[1]
	if len(key) < int(2+addrLen) {
		return nil, fmt.Errorf("invalid bank balance key: insufficient length for address")
	}

	address := string(key[2 : 2+addrLen])
	denom := string(key[2+addrLen:])

	return &StreamEvent{
		Module:      ModuleNameBank,
		EventType:   EventTypeBalanceChange,
		Address:     address,
		Denom:       denom,
		BlockHeight: 0, // Will be set by the dispatcher
	}, nil
}

// parseStakingEvent parses staking module events
func (l *StreamingListener) parseStakingEvent(key []byte, value []byte, delete bool) (*StreamEvent, error) {
	if len(key) == 0 {
		return nil, nil
	}

	prefix := key[0]

	switch {
	case prefix == StakingDelegationPrefix[0]:
		return l.parseStakingDelegationEvent(key, value, delete)
	case prefix == StakingUnbondingDelegationPrefix[0]:
		return l.parseStakingUnbondingEvent(key, value, delete)
	default:
		return nil, nil
	}
}

// parseStakingDelegationEvent parses delegation events
func (l *StreamingListener) parseStakingDelegationEvent(key []byte, value []byte, delete bool) (*StreamEvent, error) {
	// Parse delegation key format: prefix + delegator_addr_len + delegator_addr + validator_addr
	if len(key) < 2 {
		return nil, fmt.Errorf("invalid delegation key length")
	}

	delAddrLen := key[1]
	if len(key) < int(2+delAddrLen) {
		return nil, fmt.Errorf("invalid delegation key: insufficient length for delegator address")
	}

	delegatorAddr := string(key[2 : 2+delAddrLen])
	validatorAddr := string(key[2+delAddrLen:])

	return &StreamEvent{
		Module:           ModuleNameStaking,
		EventType:        EventTypeDelegationChange,
		Address:          delegatorAddr,
		SecondaryAddress: validatorAddr,
		BlockHeight:      0, // Will be set by the dispatcher
	}, nil
}

// parseStakingUnbondingEvent parses unbonding delegation events
func (l *StreamingListener) parseStakingUnbondingEvent(key []byte, value []byte, delete bool) (*StreamEvent, error) {
	// Parse unbonding delegation key format: prefix + delegator_addr_len + delegator_addr + validator_addr
	if len(key) < 2 {
		return nil, fmt.Errorf("invalid unbonding delegation key length")
	}

	delAddrLen := key[1]
	if len(key) < int(2+delAddrLen) {
		return nil, fmt.Errorf("invalid unbonding delegation key: insufficient length for delegator address")
	}

	delegatorAddr := string(key[2 : 2+delAddrLen])
	validatorAddr := string(key[2+delAddrLen:])

	return &StreamEvent{
		Module:           ModuleNameStaking,
		EventType:        EventTypeUnbondingDelegationChange,
		Address:          delegatorAddr,
		SecondaryAddress: validatorAddr,
		BlockHeight:      0, // Will be set by the dispatcher
	}, nil
}

// generateSubscriptionKey creates a subscription key from a StreamEvent
func GenerateSubscriptionKey(subscriptionType, address, secondaryAddress, denom string) SubscriptionKey {
	return SubscriptionKey{
		SubscriptionType: subscriptionType,
		Address:          address,
		SecondaryAddress: secondaryAddress,
		Denom:            denom,
	}
}
