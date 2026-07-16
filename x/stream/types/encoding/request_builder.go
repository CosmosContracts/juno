package encoding

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"google.golang.org/protobuf/runtime/protoiface"

	gogoproto "github.com/cosmos/gogoproto/proto"
)

var (
	errMethodDescriptorRequired = errors.New("method descriptor required")
	errTypeNameRequired         = errors.New("type name required")
	errTypeNotRegistered        = errors.New("type not registered")
	errTypeNotPointer           = errors.New("type is not a pointer")
	errTypeNotProtoMessage      = errors.New("type does not implement protoiface.MessageV1")
	errInvalidFieldTarget       = errors.New("field cannot accept nested value")
	errFieldNotMessage          = errors.New("field is not a message")
)

// BuildRequestMessage instantiates and populates the protobuf request expected
// by the target gRPC method using the provided field map.
func BuildRequestMessage(desc *MethodDescriptor, fields map[string]string) (protoiface.MessageV1, error) {
	if desc == nil {
		return nil, errMethodDescriptorRequired
	}

	msg, err := NewMessageByName(desc.RequestType)
	if err != nil {
		return nil, err
	}

	msgValue := reflect.ValueOf(msg)

	if len(fields) == 0 {
		return msg, nil
	}

	normalized := normalizeFieldMap(fields)
	structValue := msgValue.Elem()
	for key, value := range normalized {
		if value == "" {
			continue
		}
		if err := assignValue(structValue, key, value); err != nil {
			return nil, err
		}
	}

	return msg, nil
}

func assignValue(structValue reflect.Value, fieldKey, value string) error {
	if fieldKey == "" {
		return nil
	}
	segments := strings.Split(fieldKey, ".")
	return assignNestedValue(structValue, segments, fieldKey, value)
}

func assignNestedValue(target reflect.Value, segments []string, originalKey, value string) error {
	if len(segments) == 0 {
		return nil
	}
	fieldName := SnakeToCamel(segments[0])
	field := target.FieldByName(fieldName)
	if !field.IsValid() || !field.CanSet() {
		return nil
	}

	if len(segments) == 1 {
		return setScalarValue(field, fieldName, value)
	}

	switch field.Kind() { //nolint:exhaustive // Only pointer and struct require special handling here.
	case reflect.Ptr:
		if field.Type().Elem().Kind() != reflect.Struct {
			return fmt.Errorf("%w: %s (%s)", errInvalidFieldTarget, fieldName, originalKey)
		}
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return assignNestedValue(field.Elem(), segments[1:], originalKey, value)
	case reflect.Struct:
		return assignNestedValue(field, segments[1:], originalKey, value)
	default:
		return fmt.Errorf("%w: %s (%s)", errFieldNotMessage, fieldName, originalKey)
	}
}

func setScalarValue(field reflect.Value, fieldName, value string) error {
	if !field.CanSet() {
		return nil
	}
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		field = field.Elem()
	}

	//nolint:exhaustive // Only handle scalar kinds we expect to encounter.
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int32, reflect.Int64, reflect.Int, reflect.Uint32, reflect.Uint64, reflect.Uint:
		isUnsigned := field.Kind() == reflect.Uint32 || field.Kind() == reflect.Uint64 || field.Kind() == reflect.Uint
		if isUnsigned {
			parsed, err := strconv.ParseUint(value, 10, field.Type().Bits())
			if err != nil {
				return fmt.Errorf("parse unsigned field %s: %w", fieldName, err)
			}
			field.SetUint(parsed)
		} else {
			parsed, err := strconv.ParseInt(value, 10, field.Type().Bits())
			if err != nil {
				return fmt.Errorf("parse field %s: %w", fieldName, err)
			}
			field.SetInt(parsed)
		}
	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(value, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("parse float field %s: %w", fieldName, err)
		}
		field.SetFloat(parsed)
	case reflect.Bool:
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("parse bool field %s: %w", fieldName, err)
		}
		field.SetBool(parsed)
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.Uint8 {
			decoded, err := decodeBytes(value)
			if err != nil {
				return fmt.Errorf("parse bytes field %s: %w", fieldName, err)
			}
			field.SetBytes(decoded)
		}
	default:
		// Leave unsupported types unset.
	}
	return nil
}

func decodeBytes(value string) ([]byte, error) {
	if value == "" {
		return nil, nil
	}
	if decoded, err := base64.StdEncoding.DecodeString(value); err == nil {
		return decoded, nil
	}
	if strings.HasPrefix(value, "0x") || strings.HasPrefix(value, "0X") {
		return hex.DecodeString(value[2:])
	}
	return hex.DecodeString(value)
}

func normalizeFieldMap(fields map[string]string) map[string]string {
	if len(fields) == 0 {
		return map[string]string{}
	}
	normalized := make(map[string]string, len(fields))
	for key, value := range fields {
		canonical := normalizeFieldKey(key)
		if canonical == "" {
			continue
		}
		normalized[canonical] = value
	}
	return normalized
}

func normalizeFieldKey(key string) string {
	if key == "" {
		return ""
	}
	segments := strings.Split(key, ".")
	for i, segment := range segments {
		converted := CamelToSnake(segment)
		if converted == "" {
			converted = strings.ToLower(segment)
		}
		segments[i] = converted
	}
	return strings.Join(segments, ".")
}

// NewMessageByName returns an empty protobuf message for the provided type name.
func NewMessageByName(typeName string) (protoiface.MessageV1, error) {
	if typeName == "" {
		return nil, errTypeNameRequired
	}

	msgType := gogoproto.MessageType(typeName)
	if msgType == nil {
		return nil, fmt.Errorf("%w: %s", errTypeNotRegistered, typeName)
	}
	if msgType.Kind() != reflect.Pointer {
		return nil, fmt.Errorf("%w: %s", errTypeNotPointer, typeName)
	}

	instance := reflect.New(msgType.Elem()).Interface()
	message, ok := instance.(protoiface.MessageV1)
	if !ok {
		return nil, fmt.Errorf("%w: %s", errTypeNotProtoMessage, typeName)
	}
	return message, nil
}
