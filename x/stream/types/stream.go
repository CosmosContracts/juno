package types

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	proto "github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/log"

	"github.com/CosmosContracts/juno/v30/x/stream/types/encoding"
)

// ResolvedStream contains the metadata required to service a dynamic stream.
type ResolvedStream struct {
	Descriptor *encoding.MethodDescriptor
	Fields     map[string]string
	Key        encoding.StreamEvent
}

// RunStreamParams configures a streaming session that fans out updates from the
// subscription registry using the provided fetch and send hooks.
type RunStreamParams struct {
	// Registry manages subscriptions and connection accounting.
	Registry *SubscriptionRegistry
	// Key identifies the subscription to maintain.
	Key encoding.StreamEvent
	// ConnectionKind identifies the transport (gRPC or WebSocket).
	ConnectionKind ConnectionKind
	// ConnectionMeta carries optional metadata for connection tracking.
	ConnectionMeta ConnectionMetadata
	// Fetch is invoked to obtain the payload that should be sent to the client.
	// It is called once for the initial snapshot and again after each event.
	Fetch func(context.Context) (proto.Message, error)
	// Send delivers the payload to the downstream transport.
	Send func(proto.Message) error
	// OnEvent, when provided, is invoked for each matching StreamEvent.
	OnEvent func(encoding.StreamEvent)
	// Logger is optional and used for diagnostic messages.
	Logger log.Logger
	// OnConnect is invoked when the connection is successfully registered.
	OnConnect func(connectionID string) error
	// OnDisconnect is invoked after the connection is unregistered.
	OnDisconnect func(connectionID string)
}

// RunStream orchestrates the full lifecycle of a streaming session. It performs
// connection accounting, delivers the initial snapshot and reacts to subsequent
// stream events until the context is cancelled.
func RunStream(ctx context.Context, params RunStreamParams) error {
	logger := params.Logger
	registry := params.Registry
	kind := params.ConnectionKind
	if ctx == nil {
		ctx = context.Background()
	}

	if !registry.CanAcceptConnection(kind) {
		return ErrConnectionLimit
	}
	connectionID, err := registry.RegisterConnection(kind, params.ConnectionMeta)
	if err != nil {
		return err
	}
	if params.OnConnect != nil {
		if err := params.OnConnect(connectionID); err != nil {
			registry.UnregisterConnection(connectionID)
			return err
		}
	}
	defer func() {
		if params.OnDisconnect != nil {
			params.OnDisconnect(connectionID)
		}
	}()
	defer registry.UnregisterConnection(connectionID)

	if !registry.AddConnectionSubscription(connectionID) {
		return ErrSubscriptionLimit
	}
	defer registry.RemoveConnectionSubscription(connectionID)

	initial, err := params.Fetch(ctx)
	if err != nil {
		return err
	}
	if err := params.Send(initial); err != nil {
		return err
	}

	lastFingerprint, err := fingerprintMessage(initial)
	if err != nil {
		return err
	}

	sendCh, subscriber := registry.Subscribe(ctx, params.Key)
	defer registry.Unsubscribe(subscriber)

	keyStr := params.Key.String()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case raw, ok := <-sendCh:
			if !ok || raw == nil {
				logger.Debug("subscription channel closed", "key", SanitizeLog(keyStr))
				return nil
			}

			event, ok := raw.(encoding.StreamEvent)
			if !ok {
				logger.Debug("ignoring non stream event", "key", SanitizeLog(keyStr), "type", fmt.Sprintf("%T", raw))
				continue
			}

			if params.OnEvent != nil {
				params.OnEvent(event)
			}

			update, err := params.Fetch(ctx)
			if err != nil {
				return err
			}
			fingerprint, err := fingerprintMessage(update)
			if err != nil {
				return err
			}
			if bytes.Equal(lastFingerprint, fingerprint) {
				continue
			}
			if err := params.Send(update); err != nil {
				return err
			}
			lastFingerprint = fingerprint
		}
	}
}

func fingerprintMessage(msg proto.Message) ([]byte, error) {
	if msg == nil {
		return nil, nil
	}
	fingerprint, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return fingerprint, nil
}

// ResolveStream locates the stream definition for the provided module/method
// pair and normalises the supplied parameters.
func ResolveStream(
	registry *encoding.DynamicRegistry,
	module string,
	method string,
	params map[string]string,
) (*ResolvedStream, error) {
	moduleKey := strings.ToLower(strings.TrimSpace(module))
	methodKey := strings.ToLower(strings.TrimSpace(method))
	if moduleKey == "" || methodKey == "" {
		return nil, ErrMissingModuleOrMethod
	}

	desc, ok := registry.LookupByModuleAndStream(moduleKey, methodKey)
	if !ok || desc == nil {
		// Fallback resolution using dynamic registry search and simple aliasing.
		if alt := resolveStreamFallback(registry, moduleKey, methodKey, params); alt != nil {
			desc = alt
		} else {
			return nil, fmt.Errorf("%w: %s/%s", ErrStreamUnavailable, moduleKey, methodKey)
		}
	}

	fields, err := sanitizeParams(desc, params)
	if err != nil {
		return nil, err
	}

	return &ResolvedStream{
		Descriptor: desc,
		Fields:     fields,
		Key: encoding.StreamEvent{
			Module: desc.Module,
			Method: desc.StreamName,
			Params: cloneStringMap(fields),
		},
	}, nil
}

// resolveStreamFallback attempts to map non-canonical method names to a valid
// stream descriptor using the DynamicRegistry. It is intentionally conservative
// and prefers exact module matches and best parameter coverage.
func resolveStreamFallback(registry *encoding.DynamicRegistry, module, method string, params map[string]string) *encoding.MethodDescriptor {
	if registry == nil {
		return nil
	}

	// 1) Simple aliasing for common short-hands.
	canonical := method
	switch module {
	case "bank":
		if method == "supply" {
			canonical = "supply_of"
		}
	case "distribution":
		// Map ambiguous rewards -> choose based on presence of validator_address
		if method == "rewards" {
			if _, hasVal := params["validator_address"]; hasVal {
				canonical = "delegation_rewards"
			} else {
				canonical = "delegation_total_rewards"
			}
		}
		if method == "commission" {
			canonical = "validator_commission"
		}
		if method == "slashes" || method == "slash_events" {
			canonical = "validator_slashes"
		}
		if method == "withdraw_address" || method == "withdraw" {
			canonical = "delegator_withdraw_address"
		}
	default:
		// no aliasing available for this module
	}

	if canonical != method {
		if desc, ok := registry.LookupByModuleAndStream(module, canonical); ok && desc != nil {
			return desc
		}
	}

	// 2) Fuzzy search within module using substring match on stream name.
	var best *encoding.MethodDescriptor
	bestScore := -1

	for _, candidate := range registry.Search(method) {
		if candidate == nil || candidate.Module != module {
			continue
		}
		// Score by intersection of provided params and candidate request fields.
		fields := make(map[string]struct{}, len(candidate.RequestFields))
		for _, f := range candidate.RequestFields {
			if f.Name != "" {
				fields[f.Name] = struct{}{}
			}
			if f.JSONName != "" {
				fields[f.JSONName] = struct{}{}
			}
		}
		score := 0
		for k := range params {
			if _, ok := fields[k]; ok {
				score++
			}
		}
		// Prefer shorter request (fewer fields) when tie, to favour more general queries.
		if score > bestScore || (score == bestScore && best != nil && len(candidate.RequestFields) < len(best.RequestFields)) {
			best = candidate
			bestScore = score
		}
	}

	return best
}

func sanitizeParams(desc *encoding.MethodDescriptor, params map[string]string) (map[string]string, error) {
	if desc == nil {
		return nil, ErrDescriptorRequired
	}
	if len(params) == 0 {
		return map[string]string{}, nil
	}

	fields := make(map[string]string)
	allowed := make(map[string]struct{}, len(desc.RequestFields))
	for _, field := range desc.RequestFields {
		allowed[field.Name] = struct{}{}
		if field.JSONName != "" {
			allowed[field.JSONName] = struct{}{}
		}

		if value, ok := params[field.Name]; ok {
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				fields[field.Name] = trimmed
			}
			continue
		}

		if field.JSONName == "" {
			continue
		}
		if value, ok := params[field.JSONName]; ok {
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				fields[field.Name] = trimmed
			}
		}
	}

	for key, value := range params {
		if strings.TrimSpace(value) == "" {
			continue
		}
		if _, ok := allowed[key]; ok {
			continue
		}
		return nil, fmt.Errorf("%w: %s", ErrUnknownParameter, key)
	}

	return fields, nil
}

func cloneStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return map[string]string{}
	}

	out := make(map[string]string, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}
