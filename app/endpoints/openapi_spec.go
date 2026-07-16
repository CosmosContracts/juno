package endpoints

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"

	_ "embed"
)

//go:embed openapi.yaml
var rawOpenAPISpec []byte

var (
	specOnce  sync.Once
	cachedDoc *openAPIDocument
)

type openAPIDocument struct {
	Paths map[string]*openAPIPathItem `yaml:"paths"`
}

type openAPIPathItem struct {
	Get        *openAPIOperation  `yaml:"get"`
	Post       *openAPIOperation  `yaml:"post"`
	Put        *openAPIOperation  `yaml:"put"`
	Patch      *openAPIOperation  `yaml:"patch"`
	Delete     *openAPIOperation  `yaml:"delete"`
	Options    *openAPIOperation  `yaml:"options"`
	Head       *openAPIOperation  `yaml:"head"`
	Trace      *openAPIOperation  `yaml:"trace"`
	Parameters []OpenAPIParameter `yaml:"parameters"`
}

type openAPIOperation struct {
	OperationID string                      `yaml:"operationId"`
	Summary     string                      `yaml:"summary"`
	Tags        []string                    `yaml:"tags"`
	Parameters  []OpenAPIParameter          `yaml:"parameters"`
	Responses   map[string]*openAPIResponse `yaml:"responses"`
}

type OpenAPIParameter struct {
	Name        string         `yaml:"name"`
	In          string         `yaml:"in"`
	Required    bool           `yaml:"required"`
	Description string         `yaml:"description"`
	Schema      *openAPISchema `yaml:"schema"`
}

type openAPIResponse struct {
	Description string                       `yaml:"description"`
	Content     map[string]*openAPIMediaType `yaml:"content"`
}

type openAPIMediaType struct {
	Schema *openAPISchema `yaml:"schema"`
}

type openAPISchema struct {
	Ref    string `yaml:"$ref"`
	Type   string `yaml:"type"`
	Format string `yaml:"format"`
}

// OpenAPIEndpoint contains the subset of metadata required by downstream consumers.
type OpenAPIEndpoint struct {
	Path           string
	Method         string
	OperationID    string
	Summary        string
	Tags           []string
	Parameters     []OpenAPIParameter
	ResponseSchema string
}

// RawOpenAPISpec returns the embedded OpenAPI document bytes.
func RawOpenAPISpec() []byte {
	return rawOpenAPISpec
}

func loadDocument() (*openAPIDocument, error) {
	var errCachedSpec error
	specOnce.Do(func() {
		var parsed openAPIDocument

		if err := yaml.Unmarshal(rawOpenAPISpec, &parsed); err != nil {
			errCachedSpec = fmt.Errorf("parse openapi spec: %w", err)
			return
		}
		cachedDoc = &parsed
	})
	return cachedDoc, errCachedSpec
}

// GetOpenAPIEndpoints returns all GET endpoints defined in the OpenAPI spec.
func GetOpenAPIEndpoints() ([]OpenAPIEndpoint, error) {
	doc, err := loadDocument()
	if err != nil {
		return nil, err
	}
	if doc == nil || len(doc.Paths) == 0 {
		return nil, nil
	}
	paths := make([]string, 0, len(doc.Paths))
	for path := range doc.Paths {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	var endpoints []OpenAPIEndpoint
	for _, path := range paths {
		item := doc.Paths[path]
		if item == nil {
			continue
		}
		if ep, ok := buildEndpoint(path, http.MethodGet, item, item.Get); ok {
			endpoints = append(endpoints, ep)
		}
	}
	return endpoints, nil
}

func buildEndpoint(path, method string, item *openAPIPathItem, op *openAPIOperation) (OpenAPIEndpoint, bool) {
	if item == nil || op == nil {
		return OpenAPIEndpoint{}, false
	}
	responseSchema := primaryResponseSchema(op)
	if responseSchema == "" {
		return OpenAPIEndpoint{}, false
	}
	params := mergeParameters(item.Parameters, op.Parameters)
	return OpenAPIEndpoint{
		Path:           path,
		Method:         method,
		OperationID:    op.OperationID,
		Summary:        op.Summary,
		Tags:           append([]string(nil), op.Tags...),
		Parameters:     params,
		ResponseSchema: responseSchema,
	}, true
}

func mergeParameters(pathParams, opParams []OpenAPIParameter) []OpenAPIParameter {
	if len(pathParams) == 0 && len(opParams) == 0 {
		return nil
	}
	combined := make([]OpenAPIParameter, 0, len(pathParams)+len(opParams))
	seen := make(map[string]int, len(pathParams)+len(opParams))
	appendParam := func(p OpenAPIParameter) {
		key := strings.ToLower(p.In) + ":" + strings.ToLower(p.Name)
		if idx, exists := seen[key]; exists {
			combined[idx] = p
			return
		}
		seen[key] = len(combined)
		combined = append(combined, p)
	}
	for _, p := range pathParams {
		appendParam(p)
	}
	for _, p := range opParams {
		appendParam(p)
	}
	return combined
}

func primaryResponseSchema(op *openAPIOperation) string {
	if op == nil || len(op.Responses) == 0 {
		return ""
	}
	statusCodes := make([]string, 0, len(op.Responses))
	for code := range op.Responses {
		statusCodes = append(statusCodes, code)
	}
	sort.Strings(statusCodes)
	for _, code := range statusCodes {
		if len(code) == 0 || code[0] != '2' {
			continue
		}
		resp := op.Responses[code]
		if resp == nil {
			continue
		}
		if schema := firstSchemaRef(resp); schema != "" {
			return schema
		}
	}
	return ""
}

func firstSchemaRef(resp *openAPIResponse) string {
	if resp == nil || len(resp.Content) == 0 {
		return ""
	}
	if mt, ok := resp.Content["application/json"]; ok {
		if ref := extractSchemaRef(mt); ref != "" {
			return ref
		}
	}
	keys := make([]string, 0, len(resp.Content))
	for k := range resp.Content {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if ref := extractSchemaRef(resp.Content[key]); ref != "" {
			return ref
		}
	}
	return ""
}

func extractSchemaRef(mt *openAPIMediaType) string {
	if mt == nil || mt.Schema == nil {
		return ""
	}
	return schemaComponentName(mt.Schema.Ref)
}

func schemaComponentName(ref string) string {
	if ref == "" {
		return ""
	}
	segments := strings.Split(ref, "/")
	return segments[len(segments)-1]
}
