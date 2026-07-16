package types

import (
	"context"

	corestore "cosmossdk.io/core/store"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewTransientKVStoreService adapts a *storetypes.TransientStoreKey into a
// corestore.KVStoreService so collections can be built over a transient
// store. The SDK's runtime package (v0.53) only offers
// runtime.NewTransientStoreService, which returns the narrower
// core/store.TransientStoreService interface that collections cannot
// consume — this adapter mirrors runtime's unexported kvStoreService but
// opens the store via a transient key. The backing store resets on every
// commit, which is exactly the lifetime the dirty-delegator set needs.
func NewTransientKVStoreService(key *storetypes.TransientStoreKey) corestore.KVStoreService {
	return transientKVStoreService{key: key}
}

type transientKVStoreService struct {
	key *storetypes.TransientStoreKey
}

func (s transientKVStoreService) OpenKVStore(ctx context.Context) corestore.KVStore {
	return transientKVStore{store: sdk.UnwrapSDKContext(ctx).KVStore(s.key)}
}

// transientKVStore wraps a storetypes.KVStore into the core/store KVStore
// interface (same shape as the SDK runtime's unexported coreKVStore).
type transientKVStore struct {
	store storetypes.KVStore
}

func (s transientKVStore) Get(key []byte) ([]byte, error) { return s.store.Get(key), nil }
func (s transientKVStore) Has(key []byte) (bool, error)   { return s.store.Has(key), nil }

func (s transientKVStore) Set(key, value []byte) error {
	s.store.Set(key, value)
	return nil
}

func (s transientKVStore) Delete(key []byte) error {
	s.store.Delete(key)
	return nil
}

func (s transientKVStore) Iterator(start, end []byte) (corestore.Iterator, error) {
	return s.store.Iterator(start, end), nil
}

func (s transientKVStore) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	return s.store.ReverseIterator(start, end), nil
}
