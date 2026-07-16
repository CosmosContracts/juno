package encoding

import (
	"sort"
	"strings"
	"unicode"
)

// StreamEvent represents a state change emitted by the dispatcher. It contains
// the module, method and parameters that identify the subscription that
// produced the event.
type StreamEvent struct {
	Module string
	Method string
	Params map[string]string
}

// String returns a canonical string representation suitable for map keys.
func (e StreamEvent) String() string {
	var builder strings.Builder
	builder.Grow(32)
	_, _ = builder.WriteString(e.Module)
	_ = builder.WriteByte('/')
	_, _ = builder.WriteString(e.Method)

	if len(e.Params) == 0 {
		return builder.String()
	}

	_ = builder.WriteByte('|')
	keys := make([]string, 0, len(e.Params))
	for label := range e.Params {
		keys = append(keys, label)
	}
	sort.Strings(keys)

	for i, label := range keys {
		if i != 0 {
			_ = builder.WriteByte(',')
		}
		_, _ = builder.WriteString(label)
		_ = builder.WriteByte('=')
		_, _ = builder.WriteString(e.Params[label])
	}

	return builder.String()
}

// CamelToSnake converts CamelCase identifiers to snake_case.
func CamelToSnake(input string) string {
	if input == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(input) + 4)

	for i, r := range input {
		if unicode.IsUpper(r) {
			if i > 0 {
				_ = builder.WriteByte('_')
			}
			_, _ = builder.WriteRune(unicode.ToLower(r))
			continue
		}
		_, _ = builder.WriteRune(r)
	}
	return builder.String()
}

// SnakeToCamel converts snake_case identifiers to CamelCase.
func SnakeToCamel(input string) string {
	if input == "" {
		return ""
	}

	parts := strings.Split(input, "_")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
	}
	return strings.Join(parts, "")
}
