package protobufjson

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Unmarshal(b []byte, m proto.Message) error {
	return UnmarshalOptions{}.Unmarshal(b, m)
}

func (o UnmarshalOptions) Unmarshal(b []byte, m proto.Message) error {
	if m == nil {
		return fmt.Errorf("message is nil")
	}
	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	return o.unmarshalMessage(data, m.ProtoReflect())
}

func (o UnmarshalOptions) unmarshalMessage(data map[string]interface{}, md protoreflect.Message) error {
	fields := md.Descriptor().Fields()
	fieldByName := make(map[string]protoreflect.FieldDescriptor)
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		fieldByName[string(field.Name())] = field
		fieldByName[string(field.JSONName())] = field
	}
	
	oneofLastSet := make(map[protoreflect.OneofDescriptor]protoreflect.FieldDescriptor)
	
	for key, value := range data {
		field, ok := fieldByName[key]
		if !ok {
			if o.DiscardUnknown {
				continue
			}
			return fmt.Errorf("unknown field: %s", key)
		}
		
		if field.IsMap() {
			mapData, ok := value.(map[string]interface{})
			if !ok {
				return fmt.Errorf("expected object for map field %s", key)
			}
			if err := o.unmarshalMap(field, mapData, md); err != nil {
				return err
			}
			continue
		}
		
		if field.IsList() {
			listData, ok := value.([]interface{})
			if !ok {
				return fmt.Errorf("expected array for repeated field %s", key)
			}
			if err := o.unmarshalList(field, listData, md); err != nil {
				return err
			}
			continue
		}
		
		if field.ContainingOneof() != nil {
			oneof := field.ContainingOneof()
			oneofLastSet[oneof] = field
		}
		
		if err := o.unmarshalField(field, value, md); err != nil {
			return err
		}
	}
	
	for oneof, lastField := range oneofLastSet {
		fields := oneof.Fields()
		for i := 0; i < fields.Len(); i++ {
			field := fields.Get(i)
			if field != lastField {
				md.Clear(field)
			}
		}
	}
	
	return nil
}

func (o UnmarshalOptions) unmarshalMap(field protoreflect.FieldDescriptor, data map[string]interface{}, md protoreflect.Message) error {
	mapValue := md.Mutable(field).Map()
	keyField := field.MapKey()
	valField := field.MapValue()
	
	for keyStr, val := range data {
		mapKey, err := unmarshalMapKey(keyField, keyStr)
		if err != nil {
			return err
		}
		protoreflectValue, err := o.unmarshalValue(valField, val)
		if err != nil {
			return err
		}
		mapValue.Set(mapKey, protoreflectValue)
	}
	
	return nil
}

func unmarshalMapKey(keyField protoreflect.FieldDescriptor, keyStr string) (protoreflect.MapKey, error) {
	switch keyField.Kind() {
	case protoreflect.StringKind:
		return protoreflect.ValueOf(keyStr).MapKey(), nil
	case protoreflect.BoolKind:
		b, err := strconv.ParseBool(keyStr)
		if err != nil {
			return protoreflect.MapKey{}, err
		}
		return protoreflect.ValueOf(b).MapKey(), nil
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		i, err := strconv.ParseInt(keyStr, 10, 32)
		if err != nil {
			return protoreflect.MapKey{}, err
		}
		return protoreflect.ValueOf(int32(i)).MapKey(), nil
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		i, err := strconv.ParseInt(keyStr, 10, 64)
		if err != nil {
			return protoreflect.MapKey{}, err
		}
		return protoreflect.ValueOf(i).MapKey(), nil
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		u, err := strconv.ParseUint(keyStr, 10, 32)
		if err != nil {
			return protoreflect.MapKey{}, err
		}
		return protoreflect.ValueOf(uint32(u)).MapKey(), nil
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		u, err := strconv.ParseUint(keyStr, 10, 64)
		if err != nil {
			return protoreflect.MapKey{}, err
		}
		return protoreflect.ValueOf(u).MapKey(), nil
	default:
		return protoreflect.MapKey{}, fmt.Errorf("unsupported map key kind: %v", keyField.Kind())
	}
}

func (o UnmarshalOptions) unmarshalList(field protoreflect.FieldDescriptor, data []interface{}, md protoreflect.Message) error {
	listValue := md.Mutable(field).List()
	for _, item := range data {
		val, err := o.unmarshalValue(field, item)
		if err != nil {
			return err
		}
		listValue.Append(val)
	}
	return nil
}

func (o UnmarshalOptions) unmarshalField(field protoreflect.FieldDescriptor, value interface{}, md protoreflect.Message) error {
	if field.Kind() == protoreflect.MessageKind {
		if isWellKnownType(field.Message()) {
			protoreflectValue, err := unmarshalWellKnownType(field.Message(), value)
			if err != nil {
				return err
			}
			md.Set(field, protoreflectValue)
			return nil
		}
		msgData, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected object for message field %s", field.Name())
		}
		msgValue := md.Mutable(field).Message()
		return o.unmarshalMessage(msgData, msgValue)
	}
	
	protoreflectValue, err := o.unmarshalValue(field, value)
	if err != nil {
		return err
	}
	md.Set(field, protoreflectValue)
	return nil
}

func (o UnmarshalOptions) unmarshalValue(field protoreflect.FieldDescriptor, value interface{}) (protoreflect.Value, error) {
	switch field.Kind() {
	case protoreflect.BoolKind:
		return unmarshalBool(value)
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return unmarshalInt32(value)
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return unmarshalInt64(value)
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return unmarshalUint32(value)
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return unmarshalUint64(value)
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return unmarshalFloat(field.Kind(), value)
	case protoreflect.StringKind:
		return unmarshalString(value)
	case protoreflect.BytesKind:
		return unmarshalBytes(value)
	case protoreflect.EnumKind:
		return unmarshalEnum(field.Enum(), value)
	case protoreflect.MessageKind:
		if isWellKnownType(field.Message()) {
			return unmarshalWellKnownType(field.Message(), value)
		}
		msgData, ok := value.(map[string]interface{})
		if !ok {
			return protoreflect.Value{}, fmt.Errorf("expected object for message field")
		}
		msg := dynamicpb.NewMessage(field.Message())
		if err := o.unmarshalMessage(msgData, msg); err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfMessage(msg), nil
	default:
		return protoreflect.Value{}, fmt.Errorf("unsupported kind: %v", field.Kind())
	}
}

func unmarshalBool(value interface{}) (protoreflect.Value, error) {
	switch v := value.(type) {
	case bool:
		return protoreflect.ValueOf(v), nil
	case string:
		b, err := strconv.ParseBool(v)
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOf(b), nil
	default:
		return protoreflect.Value{}, fmt.Errorf("expected bool, got %T", value)
	}
}

func unmarshalInt32(value interface{}) (protoreflect.Value, error) {
	switch v := value.(type) {
	case float64:
		return protoreflect.ValueOf(int32(v)), nil
	case string:
		i, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOf(int32(i)), nil
	default:
		return protoreflect.Value{}, fmt.Errorf("expected int32, got %T", value)
	}
}

func unmarshalInt64(value interface{}) (protoreflect.Value, error) {
	switch v := value.(type) {
	case float64:
		return protoreflect.ValueOf(int64(v)), nil
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOf(i), nil
	default:
		return protoreflect.Value{}, fmt.Errorf("expected int64, got %T", value)
	}
}

func unmarshalUint32(value interface{}) (protoreflect.Value, error) {
	switch v := value.(type) {
	case float64:
		return protoreflect.ValueOf(uint32(v)), nil
	case string:
		u, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOf(uint32(u)), nil
	default:
		return protoreflect.Value{}, fmt.Errorf("expected uint32, got %T", value)
	}
}

func unmarshalUint64(value interface{}) (protoreflect.Value, error) {
	switch v := value.(type) {
	case float64:
		return protoreflect.ValueOf(uint64(v)), nil
	case string:
		u, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOf(u), nil
	default:
		return protoreflect.Value{}, fmt.Errorf("expected uint64, got %T", value)
	}
}

func unmarshalFloat(kind protoreflect.Kind, value interface{}) (protoreflect.Value, error) {
	switch v := value.(type) {
	case float64:
		if kind == protoreflect.FloatKind {
			return protoreflect.ValueOf(float32(v)), nil
		}
		return protoreflect.ValueOf(v), nil
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return protoreflect.Value{}, err
		}
		if kind == protoreflect.FloatKind {
			return protoreflect.ValueOf(float32(f)), nil
		}
		return protoreflect.ValueOf(f), nil
	default:
		return protoreflect.Value{}, fmt.Errorf("expected float, got %T", value)
	}
}

func unmarshalString(value interface{}) (protoreflect.Value, error) {
	switch v := value.(type) {
	case string:
		return protoreflect.ValueOf(v), nil
	default:
		return protoreflect.Value{}, fmt.Errorf("expected string, got %T", value)
	}
}

func unmarshalBytes(value interface{}) (protoreflect.Value, error) {
	switch v := value.(type) {
	case string:
		return protoreflect.ValueOf([]byte(v)), nil
	default:
		return protoreflect.Value{}, fmt.Errorf("expected bytes (string), got %T", value)
	}
}

func unmarshalEnum(enumDesc protoreflect.EnumDescriptor, value interface{}) (protoreflect.Value, error) {
	switch v := value.(type) {
	case string:
		enumValDesc := enumDesc.Values().ByName(protoreflect.Name(v))
		if enumValDesc != nil {
			return protoreflect.ValueOf(enumValDesc.Number()), nil
		}
		n, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return protoreflect.Value{}, fmt.Errorf("invalid enum value: %s", v)
		}
		return protoreflect.ValueOf(protoreflect.EnumNumber(n)), nil
	case float64:
		return protoreflect.ValueOf(protoreflect.EnumNumber(int32(v))), nil
	default:
		return protoreflect.Value{}, fmt.Errorf("expected enum (string or number), got %T", value)
	}
}

func unmarshalWellKnownType(md protoreflect.MessageDescriptor, value interface{}) (protoreflect.Value, error) {
	name := string(md.FullName())
	switch name {
	case "google.protobuf.Timestamp":
		return unmarshalTimestamp(value)
	case "google.protobuf.Duration":
		return unmarshalDuration(value)
	default:
		return protoreflect.Value{}, fmt.Errorf("unknown well-known type: %s", name)
	}
}

func unmarshalTimestamp(value interface{}) (protoreflect.Value, error) {
	s, ok := value.(string)
	if !ok {
		return protoreflect.Value{}, fmt.Errorf("expected string for Timestamp, got %T", value)
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return protoreflect.Value{}, fmt.Errorf("invalid Timestamp format: %w", err)
	}
	ts := timestamppb.New(t)
	return protoreflect.ValueOfMessage(ts.ProtoReflect()), nil
}

func unmarshalDuration(value interface{}) (protoreflect.Value, error) {
	s, ok := value.(string)
	if !ok {
		return protoreflect.Value{}, fmt.Errorf("expected string for Duration, got %T", value)
	}
	if !strings.HasSuffix(s, "s") {
		return protoreflect.Value{}, fmt.Errorf("invalid Duration format: must end with 's'")
	}
	durStr := s[:len(s)-1]
	var d time.Duration
	var err error
	
	if strings.Contains(durStr, ".") {
		parts := strings.SplitN(durStr, ".", 2)
		secs, err1 := strconv.ParseInt(parts[0], 10, 64)
		if err1 != nil {
			return protoreflect.Value{}, fmt.Errorf("invalid Duration: %w", err1)
		}
		fracStr := parts[1]
		if len(fracStr) > 9 {
			fracStr = fracStr[:9]
		} else {
			fracStr = fracStr + strings.Repeat("0", 9-len(fracStr))
		}
		nanos, err2 := strconv.ParseInt(fracStr, 10, 64)
		if err2 != nil {
			return protoreflect.Value{}, fmt.Errorf("invalid Duration: %w", err2)
		}
		d = time.Duration(secs)*time.Second + time.Duration(nanos)*time.Nanosecond
	} else {
		secs, err1 := strconv.ParseInt(durStr, 10, 64)
		if err1 != nil {
			return protoreflect.Value{}, fmt.Errorf("invalid Duration: %w", err1)
		}
		d = time.Duration(secs) * time.Second
	}
	
	_ = err
	dur := durationpb.New(d)
	return protoreflect.ValueOfMessage(dur.ProtoReflect()), nil
}
