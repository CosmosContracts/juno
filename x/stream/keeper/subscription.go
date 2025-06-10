package keeper

import (
	"context"
	"sync"

	"cosmossdk.io/log"
	"github.com/CosmosContracts/juno/v30/x/stream/types"
)

// Subscriber represents an active subscription
type Subscriber struct {
	ctx    context.Context
	sendCh chan<- any
	key    types.SubscriptionKey
}

// SubscriptionRegistry manages active subscriptions
type SubscriptionRegistry struct {
	mu          sync.RWMutex
	subscribers map[string]map[*Subscriber]bool // key -> set of subscribers
	logger      log.Logger
}

// NewSubscriptionRegistry creates a new subscription registry
func NewSubscriptionRegistry(logger log.Logger) *SubscriptionRegistry {
	return &SubscriptionRegistry{
		subscribers: make(map[string]map[*Subscriber]bool),
		logger:      logger.With("component", "subscription-registry"),
	}
}

// Subscribe adds a new subscription
func (r *SubscriptionRegistry) Subscribe(key types.SubscriptionKey, ctx context.Context, sendCh chan<- any) *Subscriber {
	r.mu.Lock()
	defer r.mu.Unlock()

	keyStr := key.String()

	// Create subscriber set if it doesn't exist
	if r.subscribers[keyStr] == nil {
		r.subscribers[keyStr] = make(map[*Subscriber]bool)
	}

	subscriber := &Subscriber{
		ctx:    ctx,
		sendCh: sendCh,
		key:    key,
	}

	r.subscribers[keyStr][subscriber] = true

	r.logger.Debug("new subscription", "key", keyStr, "total_subs", len(r.subscribers[keyStr]))

	return subscriber
}

// Unsubscribe removes a subscription
func (r *SubscriptionRegistry) Unsubscribe(subscriber *Subscriber) {
	r.mu.Lock()
	defer r.mu.Unlock()

	keyStr := subscriber.key.String()

	if subs, exists := r.subscribers[keyStr]; exists {
		delete(subs, subscriber)

		// Clean up empty sets
		if len(subs) == 0 {
			delete(r.subscribers, keyStr)
		}

		r.logger.Debug("removed subscription", "key", keyStr, "remaining_subs", len(subs))
	}
}

// FanOut distributes an event to all matching subscribers
func (r *SubscriptionRegistry) FanOut(event types.StreamEvent, data any) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Generate possible subscription keys that match this event
	keys := r.generateMatchingKeys(event)

	r.logger.Debug("fanout event",
		"event", event,
		"matching_keys", keys,
		"total_subscribers", len(r.subscribers))

	for _, keyStr := range keys {
		if subs, exists := r.subscribers[keyStr]; exists {
			r.logger.Debug("fanning out to subscribers", "key", keyStr, "count", len(subs))
			r.fanOutToSubscribers(subs, data, keyStr)
		} else {
			r.logger.Debug("no subscribers for key", "key", keyStr)
		}
	}
}

// generateMatchingKeys generates all subscription keys that could match this event
func (r *SubscriptionRegistry) generateMatchingKeys(event types.StreamEvent) []string {
	var keys []string

	switch event.Module {
	case types.ModuleNameBank:
		// Balance-specific subscription
		if event.Denom != "" {
			key := types.GenerateSubscriptionKey(types.SubscriptionTypeBalance, event.Address, "", event.Denom)
			keys = append(keys, key.String())
		}

		// All balances subscription
		key := types.GenerateSubscriptionKey(types.SubscriptionTypeAllBalances, event.Address, "", "")
		keys = append(keys, key.String())

	case types.ModuleNameStaking:
		switch event.EventType {
		case types.EventTypeDelegationChange:
			// Specific delegation subscription
			if event.SecondaryAddress != "" {
				key := types.GenerateSubscriptionKey(types.SubscriptionTypeDelegation, event.Address, event.SecondaryAddress, "")
				keys = append(keys, key.String())
			}

			// All delegations subscription
			key := types.GenerateSubscriptionKey(types.SubscriptionTypeDelegations, event.Address, "", "")
			keys = append(keys, key.String())

		case types.EventTypeUnbondingDelegationChange:
			// Specific unbonding delegation subscription
			if event.SecondaryAddress != "" {
				key := types.GenerateSubscriptionKey(types.SubscriptionTypeUnbondingDelegation, event.Address, event.SecondaryAddress, "")
				keys = append(keys, key.String())
			}

			// All unbonding delegations subscription
			key := types.GenerateSubscriptionKey(types.SubscriptionTypeUnbondingDelegations, event.Address, "", "")
			keys = append(keys, key.String())
		}
	}

	return keys
}

// fanOutToSubscribers sends data to all subscribers in the set
func (r *SubscriptionRegistry) fanOutToSubscribers(subs map[*Subscriber]bool, data any, keyStr string) {
	toRemove := make([]*Subscriber, 0)

	for sub := range subs {
		// Check if subscriber's context is still active
		select {
		case <-sub.ctx.Done():
			toRemove = append(toRemove, sub)
			continue
		default:
		}

		// Try to send data (non-blocking)
		select {
		case sub.sendCh <- data:
			// Successfully sent
		default:
			// Channel is full, mark for removal
			r.logger.Warn("subscriber channel full, removing", "key", keyStr)
			IncrementBufferOverflow(sub.key.SubscriptionType)
			toRemove = append(toRemove, sub)
		}
	}

	// Remove inactive/overflowing subscribers
	for _, sub := range toRemove {
		delete(subs, sub)
	}

	if len(toRemove) > 0 {
		r.logger.Debug("removed inactive subscribers", "count", len(toRemove), "key", keyStr)
	}
}

// GetStats returns subscription statistics
func (r *SubscriptionRegistry) GetStats() map[string]int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := make(map[string]int)
	totalSubs := 0

	for key, subs := range r.subscribers {
		count := len(subs)
		stats[key] = count
		totalSubs += count
	}

	stats["total"] = totalSubs
	return stats
}

// UpdateMetrics updates prometheus metrics for subscriptions
func (r *SubscriptionRegistry) UpdateMetrics() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Count subscriptions by type
	typeCounts := make(map[string]int)

	for _, subs := range r.subscribers {
		for sub := range subs {
			typeCounts[sub.key.SubscriptionType]++
		}
	}

	// Update metrics for each type
	for subType, count := range typeCounts {
		UpdateSubscriptionMetrics(subType, count)
	}
}

// CloseAll closes all active subscriptions
func (r *SubscriptionRegistry) CloseAll() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for key, subs := range r.subscribers {
		for sub := range subs {
			// Try to send a nil to signal close, but don't block
			select {
			case sub.sendCh <- nil:
			default:
			}
		}
		delete(r.subscribers, key)
	}

	r.logger.Info("closed all subscriptions")
}
