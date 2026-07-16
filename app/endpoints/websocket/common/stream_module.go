package common

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/mux"

	proto "github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/log"

	"github.com/CosmosContracts/juno/v30/app/endpoints"
	"github.com/CosmosContracts/juno/v30/x/stream/types"
	"github.com/CosmosContracts/juno/v30/x/stream/types/encoding"
)

const wsRoutePrefix = "/ws"

// RouteDefinition defines a WebSocket route with its pattern and handler.
type RouteDefinition struct {
	Pattern string
	Handler RouteHandler
}

// RouteHandler defines a function that handles a specific route.
type RouteHandler func(http.ResponseWriter, *http.Request)

type parameterBinding struct {
	Name          string
	CanonicalName string
	Required      bool
}

type routeMetadata struct {
	pattern     string
	endpoint    endpoints.OpenAPIEndpoint
	descriptor  *encoding.MethodDescriptor
	resolver    *descriptorResolver
	pathParams  []parameterBinding
	queryParams []parameterBinding
	logger      log.Logger
	mu          sync.RWMutex
}

// BuildWSRoutes produces websocket routes based on the OpenAPI specification.
func BuildWSRoutes(
	handler *Handler,
	registry *encoding.DynamicRegistry,
	invoker *types.RouterInvoker,
	logger log.Logger,
) []RouteDefinition {
	if handler == nil || registry == nil || invoker == nil {
		return nil
	}
	if logger == nil {
		logger = log.NewNopLogger()
	}
	routeLogger := logger.With("module", "x/stream")

	opSpecs, err := endpoints.GetOpenAPIEndpoints()
	if err != nil {
		routeLogger.Error("failed to load OpenAPI specification", "error", err)
		return nil
	}
	if len(opSpecs) == 0 {
		routeLogger.Warn("OpenAPI specification did not contain any endpoints")
		return nil
	}

	resolver := newDescriptorResolver(registry)
	var routes []RouteDefinition
	for _, endpoint := range opSpecs {
		desc := resolver.Resolve(endpoint)
		if desc == nil {
			routeLogger.Info(
				"gRPC method not yet available for endpoint; handler will retry",
				"path", endpoint.Path,
				"operation", endpoint.OperationID,
				"response_schema", endpoint.ResponseSchema,
			)
		}

		pattern := wsRoutePrefix + endpoint.Path
		pathParams, queryParams := buildParameterBindings(endpoint.Parameters)
		meta := &routeMetadata{
			pattern:     pattern,
			endpoint:    endpoint,
			descriptor:  desc,
			resolver:    resolver,
			pathParams:  pathParams,
			queryParams: queryParams,
			logger: routeLogger.With(
				"path", endpoint.Path,
				"operation", endpoint.OperationID,
				"grpc_method", grpcMethodLabel(desc),
			),
		}

		routes = append(routes, RouteDefinition{
			Pattern: pattern,
			Handler: makeRouteHandler(meta, handler, invoker),
		})
	}

	if len(routes) == 0 {
		routeLogger.Warn("no websocket routes were generated from OpenAPI spec")
	}
	return routes
}

func buildParameterBindings(params []endpoints.OpenAPIParameter) ([]parameterBinding, []parameterBinding) {
	if len(params) == 0 {
		return nil, nil
	}
	var pathParams []parameterBinding
	var queryParams []parameterBinding
	for _, param := range params {
		binding := parameterBinding{
			Name:          param.Name,
			CanonicalName: canonicalizeKey(param.Name),
			Required:      param.Required,
		}
		switch strings.ToLower(param.In) {
		case "path":
			pathParams = append(pathParams, binding)
		case "query":
			queryParams = append(queryParams, binding)
		default:
			continue
		}
	}
	return pathParams, queryParams
}

func makeRouteHandler(meta *routeMetadata, handler *Handler, invoker *types.RouterInvoker) RouteHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		desc := meta.ensureDescriptor()
		if desc == nil {
			http.Error(w, "stream definition unavailable", http.StatusServiceUnavailable)
			meta.logger.Error("stream definition unavailable; descriptor missing")
			return
		}

		fields, err := extractFields(meta, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			meta.logger.Warn("invalid websocket request", "error", err)
			return
		}

		// Validate inputs early so we can surface user-friendly errors.
		if _, err := encoding.BuildRequestMessage(desc, cloneStringMap(fields)); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			meta.logger.Warn("failed to build request message", "error", err)
			return
		}

		resolved := &types.ResolvedStream{
			Descriptor: desc,
			Fields:     fields,
			Key: encoding.StreamEvent{
				Module: desc.Module,
				Method: desc.StreamName,
				Params: cloneStringMap(fields),
			},
		}

		fetch := func(context.Context) (proto.Message, error) {
			return invoker.Execute(desc, cloneStringMap(resolved.Fields))
		}

		paramsCfg := ConnectionParams{
			Writer:   w,
			Request:  r,
			Resolved: resolved,
			Fetch:    fetch,
		}

		if err := handler.ServeConnection(paramsCfg); err != nil {
			meta.logger.Error("failed to serve websocket connection", "error", err)
		}
	}
}

func extractFields(meta *routeMetadata, r *http.Request) (map[string]string, error) {
	vars := mux.Vars(r)
	fields := make(map[string]string, len(vars)+len(r.URL.Query()))
	for _, param := range meta.pathParams {
		value, ok := vars[param.Name]
		if !ok || value == "" {
			return nil, fmt.Errorf("%w: %s", types.ErrMissingPathParameter, param.Name)
		}
		fields[param.CanonicalName] = value
	}

	queryValues := r.URL.Query()
	for key, values := range queryValues {
		if len(values) == 0 {
			continue
		}
		fields[canonicalizeKey(key)] = values[len(values)-1]
	}

	for _, param := range meta.queryParams {
		if !param.Required {
			continue
		}
		if _, ok := fields[param.CanonicalName]; !ok {
			return nil, fmt.Errorf("%w: %s", types.ErrMissingQueryParameter, param.Name)
		}
	}
	return fields, nil
}

func canonicalizeKey(name string) string {
	if name == "" {
		return ""
	}
	parts := strings.Split(name, ".")
	for i, part := range parts {
		converted := encoding.CamelToSnake(part)
		if converted == "" {
			converted = strings.ToLower(part)
		}
		parts[i] = converted
	}
	return strings.Join(parts, ".")
}

func (meta *routeMetadata) ensureDescriptor() *encoding.MethodDescriptor {
	if meta == nil {
		return nil
	}
	meta.mu.RLock()
	desc := meta.descriptor
	meta.mu.RUnlock()
	if desc != nil {
		return desc
	}
	if meta.resolver == nil {
		return nil
	}
	resolved := meta.resolver.Resolve(meta.endpoint)
	if resolved == nil {
		return nil
	}
	meta.mu.Lock()
	meta.descriptor = resolved
	meta.logger = meta.logger.With("grpc_method", grpcMethodLabel(resolved))
	meta.mu.Unlock()
	return resolved
}

func grpcMethodLabel(desc *encoding.MethodDescriptor) string {
	if desc == nil {
		return "pending"
	}
	return desc.FullMethod
}

type descriptorResolver struct {
	provider func() []*encoding.MethodDescriptor
}

func newDescriptorResolver(registry *encoding.DynamicRegistry) *descriptorResolver {
	var provider func() []*encoding.MethodDescriptor
	if registry != nil {
		provider = registry.List
	}
	return &descriptorResolver{provider: provider}
}

// newDescriptorResolverFromList is intended for testing scenarios.
func newDescriptorResolverFromList(list []*encoding.MethodDescriptor) *descriptorResolver {
	return &descriptorResolver{provider: func() []*encoding.MethodDescriptor { return list }}
}

func (r *descriptorResolver) Resolve(endpoint endpoints.OpenAPIEndpoint) *encoding.MethodDescriptor {
	if r == nil || r.provider == nil {
		return nil
	}
	candidates := r.candidatesByResponse(endpoint.ResponseSchema)
	if len(candidates) == 0 {
		return nil
	}
	return resolveDescriptorFromCandidates(endpoint, candidates)
}

func (r *descriptorResolver) candidatesByResponse(responseSchema string) []*encoding.MethodDescriptor {
	if responseSchema == "" {
		return nil
	}
	methods := r.provider()
	if len(methods) == 0 {
		return nil
	}
	key := strings.ToLower(responseSchema)
	var matches []*encoding.MethodDescriptor
	for _, desc := range methods {
		if desc == nil || desc.ResponseType == "" {
			continue
		}
		if strings.ToLower(simpleTypeName(desc.ResponseType)) == key {
			matches = append(matches, desc)
		}
	}
	return matches
}

func resolveDescriptorFromCandidates(endpoint endpoints.OpenAPIEndpoint, candidates []*encoding.MethodDescriptor) *encoding.MethodDescriptor {
	if len(candidates) == 1 {
		return candidates[0]
	}

	moduleHint := moduleHintFromEndpoint(endpoint)
	filtered := filterDescriptors(candidates, func(desc *encoding.MethodDescriptor) bool {
		if moduleHint == "" {
			return true
		}
		return strings.ToLower(desc.Module) == moduleHint
	})
	if len(filtered) == 1 {
		return filtered[0]
	}
	if len(filtered) > 0 {
		candidates = filtered
	}

	operationName := operationNameFromID(endpoint.OperationID)
	if operationName != "" {
		filtered = filterDescriptors(candidates, func(desc *encoding.MethodDescriptor) bool {
			return strings.ToLower(desc.StreamName) == operationName
		})
		if len(filtered) == 1 {
			return filtered[0]
		}
		if len(filtered) > 0 {
			candidates = filtered
		}
	}

	return selectPreferredDescriptor(candidates)
}

func filterDescriptors(candidates []*encoding.MethodDescriptor, predicate func(*encoding.MethodDescriptor) bool) []*encoding.MethodDescriptor {
	if len(candidates) == 0 {
		return nil
	}
	var filtered []*encoding.MethodDescriptor
	for _, desc := range candidates {
		if desc == nil {
			continue
		}
		if predicate(desc) {
			filtered = append(filtered, desc)
		}
	}
	return filtered
}

func selectPreferredDescriptor(candidates []*encoding.MethodDescriptor) *encoding.MethodDescriptor {
	if len(candidates) == 0 {
		return nil
	}
	best := candidates[0]
	for _, candidate := range candidates[1:] {
		best = morePreferredDescriptor(best, candidate)
	}
	return best
}

func morePreferredDescriptor(a, b *encoding.MethodDescriptor) *encoding.MethodDescriptor {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	aStable := isStableVersion(a.Version)
	bStable := isStableVersion(b.Version)
	switch {
	case !aStable && bStable:
		return b
	case aStable && !bStable:
		return a
	case b.Version > a.Version:
		return b
	default:
		return a
	}
}

func isStableVersion(version string) bool {
	if version == "" {
		return true
	}
	lower := strings.ToLower(version)
	if strings.Contains(lower, "beta") || strings.Contains(lower, "alpha") || strings.Contains(lower, "rc") {
		return false
	}
	if lower[0] != 'v' {
		return false
	}
	for _, r := range lower[1:] {
		if (r < '0' || r > '9') && (r < 'a' || r > 'z') {
			return false
		}
	}
	return true
}

func simpleTypeName(full string) string {
	if full == "" {
		return ""
	}
	segments := strings.Split(full, ".")
	return segments[len(segments)-1]
}

func moduleHintFromEndpoint(endpoint endpoints.OpenAPIEndpoint) string {
	path := strings.TrimSpace(endpoint.Path)
	if strings.HasPrefix(path, "/ibc/") || strings.HasPrefix(path, "ibc/") || strings.HasPrefix(path, "/async-icq/") {
		return "ibc"
	}
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return ""
	}
	parts := strings.Split(trimmed, "/")
	if len(parts) >= 2 {
		return strings.ToLower(parts[1])
	}
	return strings.ToLower(parts[0])
}

func operationNameFromID(operationID string) string {
	if operationID == "" {
		return ""
	}
	parts := strings.Split(operationID, "_")
	if len(parts) <= 1 {
		return strings.ToLower(operationID)
	}
	return strings.ToLower(strings.Join(parts[1:], "_"))
}

func cloneStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return map[string]string{}
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
