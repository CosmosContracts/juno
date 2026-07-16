package types

import (
	"context"
	"math"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"cosmossdk.io/log"

	"github.com/CosmosContracts/juno/v30/x/stream/types/encoding"
)

// ConnectionMetadata captures identifying information for a subscriber connection.
type ConnectionMetadata struct {
	RemoteAddr   string
	ForwardedFor string
}

// ConnectionKind identifies the transport type associated with a connection.
type ConnectionKind int

const (
	ConnectionKindWebsocket ConnectionKind = iota
	ConnectionKindGRPC
)

type connectionRecord struct {
	metadata      ConnectionMetadata
	subscriptions uint32
	createdAt     time.Time
	kind          ConnectionKind
	clientID      string
}

// Subscriber represents an active subscription
type Subscriber struct {
	doneCh <-chan struct{}
	sendCh chan<- any
	key    encoding.StreamEvent
}

// SubscriptionRegistry manages active subscriptions
type SubscriptionRegistry struct {
	mu          sync.RWMutex
	subscribers map[string]map[*Subscriber]bool // key -> set of subscribers
	methodSubs  map[string]map[*Subscriber]bool // module/method -> set of subscribers
	connections map[string]*connectionRecord    // connectionID -> connection info
	wsConns     uint32
	grpcConns   uint32
	wsMax       uint32
	grpcMax     uint32
	maxSubs     uint32
	bufferSize  uint32
	clientSubs  map[string]uint32
	logger      log.Logger
}

// NewSubscriptionRegistry creates a new subscription registry with the provided configuration.
func NewSubscriptionRegistry(logger log.Logger, cfg StreamConfig) *SubscriptionRegistry {
	bufferSize := cfg.SubscriptionBufferSize
	if bufferSize == 0 {
		bufferSize = DefaultSubscriptionBufferSize
	}

	registry := &SubscriptionRegistry{
		subscribers: make(map[string]map[*Subscriber]bool),
		methodSubs:  make(map[string]map[*Subscriber]bool),
		connections: make(map[string]*connectionRecord),
		wsMax:       cfg.WsMaxConnections,
		grpcMax:     cfg.GrpcMaxConnections,
		maxSubs:     cfg.MaxSubscriptionsPerClient,
		bufferSize:  bufferSize,
		clientSubs:  make(map[string]uint32),
		logger:      logger.With("component", "subscription-registry"),
	}
	return registry
}

// CurrentConnectionLimits returns the configured limits and whether they are unbounded.
func (r *SubscriptionRegistry) CurrentConnectionLimits() (wsMax, grpcMax, maxSubscriptionsPerConnection uint32) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.wsMax, r.grpcMax, r.maxSubs
}

// CanAcceptConnection reports whether a new connection of the provided kind can be accepted.
func (r *SubscriptionRegistry) CanAcceptConnection(kind ConnectionKind) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	switch kind {
	case ConnectionKindGRPC:
		return r.grpcConns < r.grpcMax
	case ConnectionKindWebsocket:
		return r.wsConns < r.wsMax
	default:
		return true
	}
}

func clientSubscriptionsKey(kind ConnectionKind, clientID string) string {
	return strconv.Itoa(int(kind)) + "|" + clientID
}

func deriveClientID(meta ConnectionMetadata) string {
	if meta.ForwardedFor != "" {
		parts := strings.Split(meta.ForwardedFor, ",")
		if len(parts) > 0 {
			if id := strings.TrimSpace(parts[0]); id != "" {
				return id
			}
		}
	}

	addr := strings.TrimSpace(meta.RemoteAddr)
	if addr == "" {
		return ""
	}

	if host, _, err := net.SplitHostPort(addr); err == nil {
		return host
	}

	return addr
}

func applyDelta(current uint32, delta int32) uint32 {
	if delta < 0 {
		decrement := uint32(-delta)
		if current <= decrement {
			return 0
		}
		return current - decrement
	}
	return current + uint32(delta)
}

func methodKeyFromEvent(event encoding.StreamEvent) string {
	module := strings.ToLower(strings.TrimSpace(event.Module))
	method := strings.ToLower(strings.TrimSpace(event.Method))
	if module == "" && method == "" {
		return ""
	}
	return module + "/" + method
}

// RegisterConnection registers a new connection of the provided kind and returns its UUID.
func (r *SubscriptionRegistry) RegisterConnection(kind ConnectionKind, meta ConnectionMetadata) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.limitReachedLocked(kind) {
		r.logger.Warn("connection limit reached", "kind", kind, "ws_limit", r.wsMax, "grpc_limit", r.grpcMax)
		return "", ErrConnectionLimit
	}

	clientID := deriveClientID(meta)
	connectionID := uuid.New().String()
	r.connections[connectionID] = &connectionRecord{
		metadata:  meta,
		createdAt: time.Now(),
		kind:      kind,
		clientID:  clientID,
	}
	r.incrementKindCounterLocked(kind, 1)

	r.logger.Debug("connection registered",
		"connection_id", connectionID,
		"remote_addr", SanitizeLog(meta.RemoteAddr),
		"x_forwarded_for", SanitizeLog(meta.ForwardedFor),
		"kind", kind,
		"ws_total", r.wsConns,
		"grpc_total", r.grpcConns)

	return connectionID, nil
}

// UnregisterConnection removes an active connection and updates metrics.
func (r *SubscriptionRegistry) UnregisterConnection(connectionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	record, exists := r.connections[connectionID]
	if !exists {
		return
	}

	if record.subscriptions > 0 {
		r.adjustClientSubscriptionsLocked(record, -int32(record.subscriptions))
	}
	delete(r.connections, connectionID)
	r.incrementKindCounterLocked(record.kind, -1)

	r.logger.Debug("connection unregistered",
		"connection_id", connectionID,
		"remote_addr", record.metadata.RemoteAddr,
		"kind", record.kind,
		"ws_total", r.wsConns,
		"grpc_total", r.grpcConns)
}

// CanAddConnectionSubscription checks whether adding another subscription would exceed the per-connection limit.
func (r *SubscriptionRegistry) CanAddConnectionSubscription(connectionID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	record, exists := r.connections[connectionID]
	if !exists {
		return false
	}

	if r.maxSubs == math.MaxUint32 {
		return true
	}

	if record.clientID == "" {
		return true
	}

	key := clientSubscriptionsKey(record.kind, record.clientID)
	total := r.clientSubs[key]
	return total < r.maxSubs
}

// AddConnectionSubscription increments the subscription count for a connection if under limit.
func (r *SubscriptionRegistry) AddConnectionSubscription(connectionID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	record, exists := r.connections[connectionID]
	if !exists {
		return false
	}

	if r.maxSubs != math.MaxUint32 && record.clientID != "" {
		key := clientSubscriptionsKey(record.kind, record.clientID)
		total := r.clientSubs[key]
		if total >= r.maxSubs {
			r.logger.Warn("subscription limit reached for client",
				"connection_id", connectionID,
				"remote_addr", record.metadata.RemoteAddr,
				"kind", record.kind,
				"client_id", record.clientID,
				"limit", r.maxSubs)
			return false
		}
	}

	record.subscriptions++
	r.adjustClientSubscriptionsLocked(record, 1)
	return true
}

// RemoveConnectionSubscription decrements the subscription count for a connection.
func (r *SubscriptionRegistry) RemoveConnectionSubscription(connectionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if record, exists := r.connections[connectionID]; exists && record.subscriptions > 0 {
		record.subscriptions--
		r.adjustClientSubscriptionsLocked(record, -1)
	}
}

// ConnectionStats returns the total number of connections and per-connection subscription counts.
func (r *SubscriptionRegistry) ConnectionStats() (uint32, map[string]uint32) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	totals := make(map[string]uint32, len(r.connections))
	for id, record := range r.connections {
		totals[id] = record.subscriptions
	}

	return r.wsConns + r.grpcConns, totals
}

// GetActiveConnections returns a map of active connection IDs.
func (r *SubscriptionRegistry) GetActiveConnections() map[string]bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	active := make(map[string]bool, len(r.connections))
	for id := range r.connections {
		active[id] = true
	}
	return active
}

// Subscribe adds a new subscription
func (r *SubscriptionRegistry) Subscribe(ctx context.Context, key encoding.StreamEvent) (chan any, *Subscriber) {
	r.mu.Lock()
	defer r.mu.Unlock()

	keyStr := key.String()

	// Create subscriber set if it doesn't exist
	if r.subscribers[keyStr] == nil {
		r.subscribers[keyStr] = make(map[*Subscriber]bool)
	}

	sendCh := make(chan any, r.bufferSize)

	subscriber := &Subscriber{
		doneCh: ctx.Done(),
		sendCh: sendCh,
		key:    key,
	}

	r.subscribers[keyStr][subscriber] = true

	if methodKey := methodKeyFromEvent(key); methodKey != "" {
		if r.methodSubs[methodKey] == nil {
			r.methodSubs[methodKey] = make(map[*Subscriber]bool)
		}
		r.methodSubs[methodKey][subscriber] = true
	}

	r.logger.Debug("new subscription", "key", SanitizeLog(keyStr), "total_subs", len(r.subscribers[keyStr]))

	return sendCh, subscriber
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

		r.logger.Debug("removed subscription", "key", SanitizeLog(keyStr), "remaining_subs", len(subs))
	}

	if methodKey := methodKeyFromEvent(subscriber.key); methodKey != "" {
		if subs, exists := r.methodSubs[methodKey]; exists {
			delete(subs, subscriber)
			if len(subs) == 0 {
				delete(r.methodSubs, methodKey)
			}
		}
	}
}

// FanOut distributes an event to all subscribers with a matching key.
func (r *SubscriptionRegistry) FanOut(event encoding.StreamEvent, data any) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	keyStr := event.String()
	subs, exists := r.subscribers[keyStr]
	var dispatched map[*Subscriber]bool

	if exists && len(subs) > 0 {
		r.logger.Debug("fanning out to subscribers", "key", keyStr, "count", len(subs))
		dispatched = make(map[*Subscriber]bool, len(subs))
		r.fanOutToSubscribers(subs, data, keyStr, dispatched)
	}

	if len(event.Params) == 0 {
		methodKey := methodKeyFromEvent(event)
		if methodKey != "" {
			methodSubs := r.methodSubs[methodKey]
			if len(methodSubs) > 0 {
				if dispatched == nil {
					dispatched = make(map[*Subscriber]bool)
				}
				r.logger.Debug("broadcasting method event", "module", event.Module, "method", event.Method, "count", len(methodSubs))
				r.fanOutToSubscribers(methodSubs, data, methodKey, dispatched)
				return
			}
		}
	}

	if (!exists || len(subs) == 0) && len(event.Params) != 0 {
		r.logger.Debug("fanout event with no subscribers", "key", keyStr)
	}
}

// Broadcast sends data to all active subscribers, regardless of their key.
// Temporary solution until event decoding is better supported
func (r *SubscriptionRegistry) Broadcast(data any) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.subscribers) == 0 {
		r.logger.Debug("broadcast with no subscribers")
		return
	}

	total := 0
	for keyStr, subs := range r.subscribers {
		if len(subs) == 0 {
			continue
		}
		total += len(subs)
		r.fanOutToSubscribers(subs, data, keyStr, nil)
	}
	r.logger.Debug("broadcast fanned out", "subscriber_sets", len(r.subscribers), "total_subscribers", total)
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
	r.methodSubs = make(map[string]map[*Subscriber]bool)
}

func (r *SubscriptionRegistry) adjustClientSubscriptionsLocked(record *connectionRecord, delta int32) {
	if record.clientID == "" || delta == 0 {
		return
	}

	key := clientSubscriptionsKey(record.kind, record.clientID)
	current := r.clientSubs[key]

	if delta < 0 {
		decrement := uint32(-delta)
		if current <= decrement {
			delete(r.clientSubs, key)
			return
		}
		r.clientSubs[key] = current - decrement
		return
	}

	r.clientSubs[key] = current + uint32(delta)
}

func (r *SubscriptionRegistry) limitReachedLocked(kind ConnectionKind) bool {
	switch kind {
	case ConnectionKindGRPC:
		return r.grpcConns >= r.grpcMax
	case ConnectionKindWebsocket:
		return r.wsConns >= r.wsMax
	default:
		return false
	}
}

func (r *SubscriptionRegistry) incrementKindCounterLocked(kind ConnectionKind, delta int32) {
	switch kind {
	case ConnectionKindGRPC:
		r.grpcConns = applyDelta(r.grpcConns, delta)
	case ConnectionKindWebsocket:
		r.wsConns = applyDelta(r.wsConns, delta)
	default:
		// Unknown kinds do not adjust counters.
	}
}

// fanOutToSubscribers sends data to all subscribers in the set.
func (r *SubscriptionRegistry) fanOutToSubscribers(subs map[*Subscriber]bool, data any, keyStr string, dispatched map[*Subscriber]bool) {
	toRemove := make([]*Subscriber, 0)
	droppedCount := 0

	for sub := range subs {
		if dispatched != nil {
			if dispatched[sub] {
				continue
			}
		}
		// Check if subscriber's context is still active
		select {
		case <-sub.doneCh:
			toRemove = append(toRemove, sub)
			continue
		default:
		}

		// Try to send data (non-blocking with backpressure)
		select {
		case sub.sendCh <- data:
			// Successfully sent
			if dispatched != nil {
				dispatched[sub] = true
			}
		default:
			// Channel is full - implement backpressure
			channelLen := len(sub.sendCh)
			channelCap := cap(sub.sendCh)
			fillPercent := float64(channelLen) / float64(channelCap) * 100

			if fillPercent >= 95 {
				// Channel is 80% or more full, drop the event
				r.logger.Warn("subscriber channel near capacity, dropping event",
					"key", keyStr,
					"channel_len", channelLen,
					"channel_cap", channelCap,
					"fill_percent", fillPercent)
				droppedCount++
			} else {
				// Still has some capacity, try a brief wait
				timer := time.NewTimer(5 * time.Millisecond)
				select {
				case sub.sendCh <- data:
					timer.Stop()
				case <-timer.C:
					// Still couldn't send, drop the event
					r.logger.Warn("subscriber channel full after wait, dropping event", "key", keyStr)
					droppedCount++
				}
			}
		}
	}

	// Remove inactive subscribers
	for _, sub := range toRemove {
		delete(subs, sub)
	}

	if len(toRemove) > 0 {
		r.logger.Debug("removed inactive subscribers", "count", len(toRemove), "key", keyStr)
	}

	if droppedCount > 0 {
		r.logger.Info("backpressure applied", "dropped_events", droppedCount, "key", keyStr)
	}
}
