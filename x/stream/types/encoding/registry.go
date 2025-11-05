package encoding

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/runtime/protoiface"

	gogoproto "github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/baseapp"
)

// DynamicRegistry inspects the application's gRPC query router and reflection
// registry to discover the full set of query methods that are available at
// runtime. It offers lightweight search helpers that caller code can use to
// experiment with building stream definitions on the fly.
type DynamicRegistry struct {
	mu sync.RWMutex

	methods        map[string]*MethodDescriptor // keyed by fully-qualified method (e.g. /cosmos.bank.v1beta1.Query/Balance)
	requestIndex   map[string]*MethodDescriptor // keyed by request message full name
	moduleIndex    map[string]map[string]*MethodDescriptor
	orderedMethods []*MethodDescriptor
}

type hybridHandlerLookup interface {
	HybridHandlerByRequestName(name string) []func(ctx context.Context, req, resp protoiface.MessageV1) error
}

// MethodDescriptor captures the key metadata required to reason about a
// discovered gRPC query. It intentionally avoids any stream-specific behaviour
// so higher layers can decide how to turn methods into subscription
// definitions.
type MethodDescriptor struct {
	// FullMethod is the canonical gRPC method path (/package.Service/Method)
	FullMethod string
	// Method is the RPC name (e.g. Balance)
	Method string
	// Module is the short module identifier derived from the service namespace (e.g. bank).
	Module string
	// Version is best-effort metadata extracted from the service namespace
	// (e.g. v1beta1). It may be empty if no obvious version segment exists
	Version string
	// RequestType is the fully-qualified name of the request message
	RequestType string
	// ResponseType is the fully-qualified name of the response message
	ResponseType string
	// RequestFields enumerates the request message fields in declaration order
	RequestFields []FieldDescriptor
	// StreamName is the canonical name used by client APIs (typically snake_case)
	StreamName string
}

// FieldDescriptor models a single request field in a discovered RPC method
type FieldDescriptor struct {
	Name        string
	JSONName    string
	Kind        string
	Cardinality string
	TypeName    string
}

var (
	errBaseAppNil         = errors.New("baseapp cannot be nil")
	errProtoRegistryNil   = errors.New("proto files registry is nil")
	errDynamicRegistryNil = errors.New("dynamic registry not initialised")
)

// NewDynamicRegistry constructs a dynamic registry for the provided BaseApp. It
// immediately scans all registered gRPC services and keeps the discovered
// metadata in memory
func NewDynamicRegistry(app *baseapp.BaseApp) (*DynamicRegistry, error) {
	if app == nil {
		return nil, errBaseAppNil
	}

	protoFiles, err := gogoproto.MergedRegistry()
	if err != nil {
		return nil, fmt.Errorf("load merged proto registry: %w", err)
	}

	reg := &DynamicRegistry{
		methods:      map[string]*MethodDescriptor{},
		requestIndex: map[string]*MethodDescriptor{},
		moduleIndex:  map[string]map[string]*MethodDescriptor{},
	}
	if err := reg.refresh(app.GRPCQueryRouter(), protoFiles); err != nil {
		return nil, err
	}
	return reg, nil
}

// NewFilesRegistry builds a dynamic registry using the provided file set. This
// constructor does not require access to the application's router and therefore
// does not filter methods by handler availability.
func NewFilesRegistry(files *protoregistry.Files) (*DynamicRegistry, error) {
	if files == nil {
		return nil, errProtoRegistryNil
	}

	reg := &DynamicRegistry{
		methods:      map[string]*MethodDescriptor{},
		requestIndex: map[string]*MethodDescriptor{},
		moduleIndex:  map[string]map[string]*MethodDescriptor{},
	}
	if err := reg.refresh(nil, files); err != nil {
		return nil, err
	}
	return reg, nil
}

// Refresh rescans the router using the supplied app. It replaces the cached
// metadata atomically so callers can safely share a registry across updates.
func (r *DynamicRegistry) Refresh(app *baseapp.BaseApp) error {
	if r == nil {
		return errDynamicRegistryNil
	}
	if app == nil {
		return errBaseAppNil
	}

	protoFiles, err := gogoproto.MergedRegistry()
	if err != nil {
		return fmt.Errorf("load merged proto registry: %w", err)
	}
	return r.refresh(app.GRPCQueryRouter(), protoFiles)
}

// List returns all discovered methods sorted lexicographically by their full
// gRPC method path.
func (r *DynamicRegistry) List() []*MethodDescriptor {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*MethodDescriptor, len(r.orderedMethods))
	copy(out, r.orderedMethods)
	return out
}

// LookupByFullMethod returns the descriptor for the provided gRPC method path.
func (r *DynamicRegistry) LookupByFullMethod(fullMethod string) (*MethodDescriptor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	desc, ok := r.methods[fullMethod]
	return desc, ok
}

// LookupByRequest returns the descriptor associated with the given request type
// if a handler exists.
func (r *DynamicRegistry) LookupByRequest(requestType string) (*MethodDescriptor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	desc, ok := r.requestIndex[requestType]
	return desc, ok
}

// LookupByModuleAndStream attempts to resolve a method descriptor using the
// provided module and stream identifiers. Both lookups are case-insensitive.
func (r *DynamicRegistry) LookupByModuleAndStream(module, stream string) (*MethodDescriptor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.moduleIndex) == 0 {
		return nil, false
	}

	moduleKey := strings.ToLower(module)
	streamKey := strings.ToLower(stream)

	methods, ok := r.moduleIndex[moduleKey]
	if !ok {
		return nil, false
	}
	desc, ok := methods[streamKey]
	return desc, ok
}

// Search performs a case-insensitive substring search across the method path,
// service namespace, request type and response type. It returns the matching
// descriptors in lexical order.
func (r *DynamicRegistry) Search(term string) []*MethodDescriptor {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if term == "" {
		out := make([]*MethodDescriptor, len(r.orderedMethods))
		copy(out, r.orderedMethods)
		return out
	}

	lowerTerm := strings.ToLower(term)
	var matches []*MethodDescriptor
	for _, desc := range r.orderedMethods {
		if desc.matches(lowerTerm) {
			matches = append(matches, desc)
		}
	}
	return matches
}

func (r *DynamicRegistry) refresh(router hybridHandlerLookup, files *protoregistry.Files) error {
	if files == nil {
		return errProtoRegistryNil
	}

	var (
		methods      []*MethodDescriptor
		seenFullPath = map[string]struct{}{}
	)

	files.RangeFiles(func(file protoreflect.FileDescriptor) bool {
		services := file.Services()
		if services.Len() == 0 {
			return true
		}

		for i := 0; i < services.Len(); i++ {
			service := services.Get(i)

			for j := 0; j < service.Methods().Len(); j++ {
				method := service.Methods().Get(j)
				requestName := string(method.Input().FullName())

				if router != nil && len(router.HybridHandlerByRequestName(requestName)) == 0 {
					continue
				}

				desc := buildMethodDescriptor(service, method)
				if _, exists := seenFullPath[desc.FullMethod]; exists {
					continue
				}
				seenFullPath[desc.FullMethod] = struct{}{}
				methods = append(methods, desc)
			}
		}

		return true
	})

	sort.Slice(methods, func(i, j int) bool {
		return methods[i].FullMethod < methods[j].FullMethod
	})

	methodIndex := make(map[string]*MethodDescriptor, len(methods))
	requestIndex := make(map[string]*MethodDescriptor, len(methods))
	moduleIndex := make(map[string]map[string]*MethodDescriptor, len(methods))
	for _, m := range methods {
		methodIndex[m.FullMethod] = m
		requestIndex[m.RequestType] = m

		moduleKey := strings.ToLower(m.Module)
		if moduleIndex[moduleKey] == nil {
			moduleIndex[moduleKey] = make(map[string]*MethodDescriptor)
		}
		streamKey := strings.ToLower(m.StreamName)
		existing := moduleIndex[moduleKey][streamKey]
		moduleIndex[moduleKey][streamKey] = preferDescriptor(existing, m)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.methods = methodIndex
	r.requestIndex = requestIndex
	r.moduleIndex = moduleIndex
	r.orderedMethods = methods
	return nil
}

func buildMethodDescriptor(
	service protoreflect.ServiceDescriptor,
	method protoreflect.MethodDescriptor,
) *MethodDescriptor {
	fullService := string(service.FullName())
	fullMethod := fmt.Sprintf("/%s/%s", fullService, method.Name())

	module, version := inferModuleAndVersion(fullService)
	requestFields := extractRequestFields(method.Input())
	streamName := deriveStreamName(module, method.Name())

	return &MethodDescriptor{
		FullMethod:    fullMethod,
		Method:        string(method.Name()),
		Module:        shortModuleName(module),
		Version:       version,
		RequestType:   string(method.Input().FullName()),
		ResponseType:  string(method.Output().FullName()),
		RequestFields: requestFields,
		StreamName:    streamName,
	}
}

func extractRequestFields(descriptor protoreflect.MessageDescriptor) []FieldDescriptor {
	if descriptor == nil {
		return nil
	}

	fields := make([]FieldDescriptor, 0, descriptor.Fields().Len())
	for i := 0; i < descriptor.Fields().Len(); i++ {
		field := descriptor.Fields().Get(i)

		fd := FieldDescriptor{
			Name:        string(field.Name()),
			JSONName:    field.JSONName(),
			Kind:        field.Kind().String(),
			Cardinality: field.Cardinality().String(),
		}

		//nolint:exhaustive // Only specific kinds require special handling.
		switch field.Kind() {
		case protoreflect.EnumKind:
			if enum := field.Enum(); enum != nil {
				fd.TypeName = string(enum.FullName())
			}
		case protoreflect.MessageKind, protoreflect.GroupKind:
			if msg := field.Message(); msg != nil {
				fd.TypeName = string(msg.FullName())
			}
		default:
			fd.TypeName = field.Kind().String()
		}

		fields = append(fields, fd)
	}
	return fields
}

func inferModuleAndVersion(serviceFullName string) (module string, version string) {
	if serviceFullName == "" {
		return "", ""
	}

	segments := strings.Split(serviceFullName, ".")
	if len(segments) <= 1 {
		return strings.ToLower(serviceFullName), ""
	}

	// The final segment is the service name (e.g. Query). Treat the preceding
	// element as a version if it matches the expected pattern.
	candidateVersion := segments[len(segments)-2]
	if isVersionSegment(candidateVersion) {
		version = candidateVersion
		segments = segments[:len(segments)-2]
	} else {
		segments = segments[:len(segments)-1]
	}

	if len(segments) == 0 {
		return "", version
	}

	module = strings.ToLower(strings.Join(segments, "."))
	return module, version
}

func isVersionSegment(segment string) bool {
	if segment == "" {
		return false
	}
	lower := strings.ToLower(segment)
	if lower[0] != 'v' {
		return false
	}
	// Accept common Cosmos SDK patterns such as v1, v1beta1, v1alpha1, etc.
	for _, r := range lower[1:] {
		if (r < '0' || r > '9') && (r < 'a' || r > 'z') {
			return false
		}
	}
	return true
}

func shortModuleName(module string) string {
	if module == "" {
		return ""
	}

	parts := strings.Split(module, ".")
	return strings.ToLower(parts[len(parts)-1])
}

func deriveStreamName(module string, method protoreflect.Name) string {
	base := CamelToSnake(string(method))

	moduleShort := shortModuleName(module)
	if moduleShort == "staking" && strings.HasPrefix(base, "delegator_") {
		return strings.TrimPrefix(base, "delegator_")
	}

	return base
}

func (m *MethodDescriptor) matches(term string) bool {
	if term == "" {
		return true
	}

	candidates := []string{
		m.FullMethod,
		m.Module,
		m.StreamName,
		m.RequestType,
		m.ResponseType,
	}

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if strings.Contains(strings.ToLower(candidate), term) {
			return true
		}
	}
	return false
}

func preferDescriptor(existing, candidate *MethodDescriptor) *MethodDescriptor {
	if existing == nil {
		return candidate
	}
	if candidate == nil {
		return existing
	}

	exStable := isStableVersion(existing.Version)
	candStable := isStableVersion(candidate.Version)

	switch {
	case !exStable && candStable:
		return candidate
	case exStable && !candStable:
		return existing
	case existing.Version == candidate.Version:
		return existing
	case candidate.Version > existing.Version:
		return candidate
	default:
		return existing
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
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
