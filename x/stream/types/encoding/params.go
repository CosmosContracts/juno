package encoding

import (
	"strconv"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// ExtractParams collects scalar request field values present on the provided message.
func ExtractParams(desc *MethodDescriptor, msg protoreflect.Message) (map[string]string, error) {
	params := make(map[string]string, len(desc.RequestFields))
	fields := msg.Descriptor().Fields()

	for _, field := range desc.RequestFields {
		fd := fields.ByName(protoreflect.Name(field.Name))
		if fd == nil && field.JSONName != "" {
			fd = fields.ByJSONName(field.JSONName)
		}
		if fd == nil || fd.IsList() || fd.IsMap() {
			continue
		}
		if !msg.Has(fd) {
			continue
		}

		if str, ok := stringifyScalarValue(fd, msg.Get(fd)); ok && str != "" {
			params[field.Name] = str
		}
	}

	return params, nil
}

func stringifyScalarValue(fd protoreflect.FieldDescriptor, value protoreflect.Value) (string, bool) {
	//nolint:exhaustive // Only scalar kinds we support need explicit handling here.
	switch fd.Kind() {
	case protoreflect.StringKind:
		if value.String() == "" {
			return "", false
		}
		return value.String(), true
	case protoreflect.BoolKind:
		return strconv.FormatBool(value.Bool()), true
	case protoreflect.EnumKind:
		return strconv.FormatInt(int64(value.Enum()), 10), true
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return strconv.FormatInt(value.Int(), 10), true
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind,
		protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return strconv.FormatUint(value.Uint(), 10), true
	default:
		return "", false
	}
}
