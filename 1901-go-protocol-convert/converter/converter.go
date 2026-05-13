package converter

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

type Converter struct{}

func NewConverter() *Converter {
	return &Converter{}
}

func (c *Converter) JSONToProto(
	jsonData map[string]interface{},
	fieldMappings map[string]string,
	msgDesc protoreflect.MessageDescriptor,
) (*dynamicpb.Message, error) {
	msg := dynamicpb.NewMessage(msgDesc)
	err := c.MapJSONToProto(jsonData, msg, fieldMappings)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (c *Converter) ProtoToJSON(
	msg *dynamicpb.Message,
	reverseMappings map[string]string,
) (map[string]interface{}, error) {
	return c.MapProtoToJSON(msg, reverseMappings), nil
}

func (c *Converter) MapJSONToProto(jsonData map[string]interface{}, msg *dynamicpb.Message, mappings map[string]string) error {
	fields := msg.Descriptor().Fields()
	for jsonField, protoField := range mappings {
		value, ok := jsonData[jsonField]
		if !ok {
			continue
		}
		field := fields.ByName(protoreflect.Name(protoField))
		if field == nil {
			return fmt.Errorf("proto field not found: %s", protoField)
		}
		if err := c.setFieldValue(msg, field, value); err != nil {
			return err
		}
	}
	return nil
}

func (c *Converter) MapProtoToJSON(msg *dynamicpb.Message, reverseMappings map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	fields := msg.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		protoFieldName := string(field.Name())
		jsonFieldName, ok := reverseMappings[protoFieldName]
		if !ok {
			jsonFieldName = protoFieldName
		}
		if field.IsList() {
			listValue := msg.Get(field).List()
			arr := make([]interface{}, 0, listValue.Len())
			for j := 0; j < listValue.Len(); j++ {
				item := listValue.Get(j)
				arr = append(arr, c.convertProtoValueToJSON(item, field))
			}
			result[jsonFieldName] = arr
		} else if field.Message() != nil {
			nestedMsg := msg.Get(field).Message().Interface()
			dynamicNestedMsg, ok := nestedMsg.(*dynamicpb.Message)
			if ok {
				result[jsonFieldName] = c.MapProtoToJSON(dynamicNestedMsg, reverseMappings)
			}
		} else {
			result[jsonFieldName] = c.convertProtoValueToJSON(msg.Get(field), field)
		}
	}
	return result
}

func (c *Converter) setFieldValue(msg *dynamicpb.Message, field protoreflect.FieldDescriptor, value interface{}) error {
	if field.IsList() {
		arr, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("expected array for repeated field %s", field.Name())
		}
		list := msg.Mutable(field).List()
		for _, item := range arr {
			protoVal, err := c.convertJSONToProtoValue(item, field)
			if err != nil {
				return err
			}
			list.Append(protoVal)
		}
		return nil
	}
	if field.Message() != nil {
		nestedData, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected object for message field %s", field.Name())
		}
		nestedMsg := msg.Mutable(field).Message()
		dynamicNestedMsg, ok := nestedMsg.(*dynamicpb.Message)
		if !ok {
			return fmt.Errorf("failed to get dynamic message for field %s", field.Name())
		}
		nestedFields := make(map[string]string)
		for k := range nestedData {
			nestedFields[k] = k
		}
		return c.MapJSONToProto(nestedData, dynamicNestedMsg, nestedFields)
	}
	protoVal, err := c.convertJSONToProtoValue(value, field)
	if err != nil {
		return err
	}
	msg.Set(field, protoVal)
	return nil
}

func (c *Converter) convertJSONToProtoValue(value interface{}, field protoreflect.FieldDescriptor) (protoreflect.Value, error) {
	switch field.Kind() {
	case protoreflect.BoolKind:
		b, ok := value.(bool)
		if !ok {
			return protoreflect.Value{}, fmt.Errorf("expected bool for field %s", field.Name())
		}
		return protoreflect.ValueOfBool(b), nil
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		var i int32
		switch v := value.(type) {
		case float64:
			i = int32(v)
		case int:
			i = int32(v)
		default:
			return protoreflect.Value{}, fmt.Errorf("expected int32 for field %s", field.Name())
		}
		return protoreflect.ValueOfInt32(i), nil
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		var i int64
		switch v := value.(type) {
		case float64:
			i = int64(v)
		case int:
			i = int64(v)
		default:
			return protoreflect.Value{}, fmt.Errorf("expected int64 for field %s", field.Name())
		}
		return protoreflect.ValueOfInt64(i), nil
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		var i uint32
		switch v := value.(type) {
		case float64:
			i = uint32(v)
		case int:
			i = uint32(v)
		default:
			return protoreflect.Value{}, fmt.Errorf("expected uint32 for field %s", field.Name())
		}
		return protoreflect.ValueOfUint32(i), nil
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		var i uint64
		switch v := value.(type) {
		case float64:
			i = uint64(v)
		case int:
			i = uint64(v)
		default:
			return protoreflect.Value{}, fmt.Errorf("expected uint64 for field %s", field.Name())
		}
		return protoreflect.ValueOfUint64(i), nil
	case protoreflect.FloatKind:
		var f float32
		switch v := value.(type) {
		case float64:
			f = float32(v)
		default:
			return protoreflect.Value{}, fmt.Errorf("expected float32 for field %s", field.Name())
		}
		return protoreflect.ValueOfFloat32(f), nil
	case protoreflect.DoubleKind:
		f, ok := value.(float64)
		if !ok {
			return protoreflect.Value{}, fmt.Errorf("expected float64 for field %s", field.Name())
		}
		return protoreflect.ValueOfFloat64(f), nil
	case protoreflect.StringKind:
		s, ok := value.(string)
		if !ok {
			return protoreflect.Value{}, fmt.Errorf("expected string for field %s", field.Name())
		}
		return protoreflect.ValueOfString(s), nil
	case protoreflect.BytesKind:
		var b []byte
		switch v := value.(type) {
		case string:
			b = []byte(v)
		case []byte:
			b = v
		default:
			return protoreflect.Value{}, fmt.Errorf("expected bytes/string for field %s", field.Name())
		}
		return protoreflect.ValueOfBytes(b), nil
	default:
		return protoreflect.Value{}, fmt.Errorf("unsupported field kind: %s", field.Kind())
	}
}

func (c *Converter) convertProtoValueToJSON(value protoreflect.Value, field protoreflect.FieldDescriptor) interface{} {
	switch field.Kind() {
	case protoreflect.BoolKind:
		return value.Bool()
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return value.Int()
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return value.Int()
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return value.Uint()
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return value.Uint()
	case protoreflect.FloatKind:
		return value.Float()
	case protoreflect.DoubleKind:
		return value.Float()
	case protoreflect.StringKind:
		return value.String()
	case protoreflect.BytesKind:
		return string(value.Bytes())
	default:
		return value.Interface()
	}
}
