package common

import (
	"context"

	"github.com/CosmosContracts/juno/v30/x/stream/types"
)

// SubscriptionRegistryAdapter adapts the types.SubscriptionRegistry to common.SubscriptionRegistry
type SubscriptionRegistryAdapter struct {
	registry *types.SubscriptionRegistry
}

// NewSubscriptionRegistryAdapter creates a new subscription registry adapter
func NewSubscriptionRegistryAdapter(registry *types.SubscriptionRegistry) SubscriptionRegistry {
	return &SubscriptionRegistryAdapter{registry: registry}
}

// Subscribe adapts the Subscribe method
func (s *SubscriptionRegistryAdapter) Subscribe(key SubscriptionKey, ctx context.Context, sendCh chan<- any) Subscriber {
	// Convert the interface key to the concrete type
	if keyImpl, ok := key.(*types.SubscriptionKey); ok {
		return &SubscriberAdapter{subscriber: s.registry.Subscribe(*keyImpl, ctx, sendCh)}
	}
	// If not the expected type, create a string-based key
	stringKey := types.SubscriptionKey{
		SubscriptionType: types.SubscriptionTypeBalance, // Default type
		Address:          key.String(),
	}
	return &SubscriberAdapter{subscriber: s.registry.Subscribe(stringKey, ctx, sendCh)}
}

// Unsubscribe adapts the Unsubscribe method
func (s *SubscriptionRegistryAdapter) Unsubscribe(subscriber Subscriber) {
	if subAdapter, ok := subscriber.(*SubscriberAdapter); ok {
		s.registry.Unsubscribe(subAdapter.subscriber)
	}
}

// SubscriberAdapter adapts the types.Subscriber to common.Subscriber
type SubscriberAdapter struct {
	subscriber *types.Subscriber
}

// SubscriptionKeyAdapter adapts types.SubscriptionKey to common.SubscriptionKey
type SubscriptionKeyAdapter struct {
	key types.SubscriptionKey
}

// NewSubscriptionKeyAdapter creates a new subscription key adapter
func NewSubscriptionKeyAdapter(key types.SubscriptionKey) SubscriptionKey {
	return &SubscriptionKeyAdapter{key: key}
}

// String returns the string representation
func (s *SubscriptionKeyAdapter) String() string {
	return s.key.String()
}