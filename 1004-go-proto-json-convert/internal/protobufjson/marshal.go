package protobufjson

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Marshal(m proto.Message) ([]byte, error) {
	return MarshalOptions{}.Marshal(m)
}

func (o MarshalOptions) Marshal(m proto.Message) ([]byte, error) {
	if m == nil {
		return nil, fmt.Errorf("message is nil")
	}
	result, err := o.marshalMessage(m.ProtoReflect())
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (o MarshalOptions) marshalMessage(md protoreflect.Message) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	fields := md.Descriptor().Fields()
	
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		fieldName := string(field.Name())
		
		if field.HasOptionalKeyword() {
			if !md.Has(field) {
				continue
			}
		}
		
		if field.IsMap() {
			mapValue := md.Get(field).Map()
			if mapValue.Len() == 0 && !field.HasOptionalKeyword() {
				result[fieldName] = map[string]interface{}{}
				continue
			}
			mapResult, err := o.marshalMap(field, mapValue)
			if err != nil {
				return nil, err
			}
			result[fieldName] = mapResult
			continue
		}
		
		if field.IsList() {
			listValue := md.Get(field).List()
			listResult, err := o.marshalList(field, listValue)
			if err != nil {
				return nil, err
			}
			result[fieldName] = listResult
			continue
		}
		
		if field.ContainingOneof() != nil {
			oneof := field.ContainingOneof()
			oneofField := md.WhichOneof(oneof)
			if oneofField == nil || oneofField != field {
				continue
			}
		}
		
		if field.Kind() == protoreflect.MessageKind && !md.Has(field) {
			continue
		}
		
		value := md.Get(field)
		fieldValue, err := o.marshalField(field, value)
		if err != nil {
			return nil, err
		}
		result[fieldName] = fieldValue
	}
	
	return result, nil
}

func (o MarshalOptions) marshalMap(field protoreflect.FieldDescriptor, mv protoreflect.Map) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	var keys []string
	keyMap := make(map[string]protoreflect.MapKey)
	
	mv.Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
		keyStr, err := marshalMapKey(field.MapKey(), k)
		if err != nil {
			return false
		}
		keys = append(keys, keyStr)
		keyMap[keyStr] = k
		return true
	})
	
	sort.Strings(keys)
	
	for _, keyStr := range keys {
		k := keyMap[keyStr]
		v := mv.Get(k)
		val, err := o.marshalValue(field.MapValue(), v)
		if err != nil {
			return nil, err
		}
		result[keyStr] = val
	}
	
	return result, nil
}

func marshalMapKey(keyField protoreflect.FieldDescriptor, k protoreflect.MapKey) (string, error) {
	switch keyField.Kind() {
	case protoreflect.StringKind:
		return k.String(), nil
	case protoreflect.BoolKind:
		return fmt.Sprintf("%t", k.Bool()), nil
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return fmt.Sprintf("%d", k.Int()), nil
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return fmt.Sprintf("%d", k.Int()), nil
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return fmt.Sprintf("%d", k.Uint()), nil
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return fmt.Sprintf("%d", k.Uint()), nil
	default:
		return "", fmt.Errorf("unsupported map key kind: %v", keyField.Kind())
	}
}

func (o MarshalOptions) marshalList(field protoreflect.FieldDescriptor, lv protoreflect.List) ([]interface{}, error) {
	result := make([]interface{}, 0, lv.Len())
	for i := 0; i < lv.Len(); i++ {
		v := lv.Get(i)
		val, err := o.marshalValue(field, v)
		if err != nil {
			return nil, err
		}
		result = append(result, val)
	}
	return result, nil
}

func (o MarshalOptions) marshalField(field protoreflect.FieldDescriptor, v protoreflect.Value) (interface{}, error) {
	if field.Kind() == protoreflect.MessageKind {
		msgVal := v.Message()
		if isWellKnownType(field.Message()) {
			return marshalWellKnownType(field.Message(), msgVal)
		}
		return o.marshalMessage(msgVal)
	}
	return o.marshalValue(field, v)
}

func (o MarshalOptions) marshalValue(field protoreflect.FieldDescriptor, v protoreflect.Value) (interface{}, error) {
	switch field.Kind() {
	case protoreflect.BoolKind:
		return v.Bool(), nil
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return v.Int(), nil
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return fmt.Sprintf("%d", v.Int()), nil
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return v.Uint(), nil
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return fmt.Sprintf("%d", v.Uint()), nil
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return v.Float(), nil
	case protoreflect.StringKind:
		return v.String(), nil
	case protoreflect.BytesKind:
		return v.Bytes(), nil
	case protoreflect.EnumKind:
		enumVal := v.Enum()
		enumDesc := field.Enum()
		enumValDesc := enumDesc.Values().ByNumber(enumVal)
		if enumValDesc != nil {
			return string(enumValDesc.Name()), nil
		}
		return int32(enumVal), nil
	case protoreflect.MessageKind:
		msg := v.Message()
		if isWellKnownType(field.Message()) {
			return marshalWellKnownType(field.Message(), msg)
		}
		return o.marshalMessage(msg)
	default:
		return nil, fmt.Errorf("unsupported kind: %v", field.Kind())
	}
}

func isWellKnownType(md protoreflect.MessageDescriptor) bool {
	name := string(md.FullName())
	return name == "google.protobuf.Timestamp" || name == "google.protobuf.Duration"
}

func marshalWellKnownType(md protoreflect.MessageDescriptor, msg protoreflect.Message) (interface{}, error) {
	name := string(md.FullName())
	switch name {
	case "google.protobuf.Timestamp":
		ts := &timestamppb.Timestamp{}
		proto.Merge(ts, msg.Interface())
		t := ts.AsTime()
		return t.Format(time.RFC3339Nano), nil
	case "google.protobuf.Duration":
		dur := &durationpb.Duration{}
		proto.Merge(dur, msg.Interface())
		d := dur.AsDuration()
		secs := int64(d.Seconds())
		nanos := d.Nanoseconds() % 1e9
		if nanos == 0 {
			return fmt.Sprintf("%ds", secs), nil
		}
		return fmt.Sprintf("%d.%09ds", secs, nanos), nil
	default:
		return nil, fmt.Errorf("unknown well-known type: %s", name)
	}
}

var _ = reflect.TypeOf
var _ = strings.Builder{}
