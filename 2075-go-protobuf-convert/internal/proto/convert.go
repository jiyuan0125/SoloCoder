package proto

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

type ConvertOptions struct {
	EmitDefaultValues bool
	PrettyPrint       bool
}

type Warning struct {
	Field   string
	Message string
}

type ConvertResult struct {
	Data     []byte
	Warnings []Warning
}

type ParseErrorResult struct {
	ParsedFields map[string]interface{}
	ErrorOffset  int
	ErrorMessage string
}

func ProtoToJSON(msg *dynamicpb.Message, opts ConvertOptions) (*ConvertResult, error) {
	warnings := make([]Warning, 0)

	checkTimestamps(msg, "", &warnings)

	marshaler := protojson.MarshalOptions{
		EmitUnpopulated: opts.EmitDefaultValues,
		UseProtoNames:   false,
		Multiline:       opts.PrettyPrint,
		Indent:          "  ",
	}

	data, err := marshaler.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal proto to JSON: %w", err)
	}

	return &ConvertResult{
		Data:     data,
		Warnings: warnings,
	}, nil
}

func JSONToProto(data []byte, md protoreflect.MessageDescriptor, opts ConvertOptions) (*dynamicpb.Message, error) {
	msg := dynamicpb.NewMessage(md)

	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: false,
	}

	if err := unmarshaler.Unmarshal(data, msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to proto: %w", err)
	}

	return msg, nil
}

func UnmarshalProtoBinary(data []byte, msg *dynamicpb.Message) (*ParseErrorResult, error) {
	err := proto.Unmarshal(data, msg)
	if err == nil {
		return nil, nil
	}

	partialParsed := make(map[string]interface{})
	errorOffset := 0
	errMsg := err.Error()

	partialMsg := proto.Clone(msg)
	mergeErr := proto.UnmarshalOptions{
		AllowPartial: true,
	}.Unmarshal(data, partialMsg)

	if mergeErr == nil {
		marshaler := protojson.MarshalOptions{
			EmitUnpopulated: false,
		}
		jsonData, jsonErr := marshaler.Marshal(partialMsg)
		if jsonErr == nil {
			var parsed map[string]interface{}
			if json.Unmarshal(jsonData, &parsed) == nil {
				partialParsed = parsed
			}
		}
	}

	return &ParseErrorResult{
		ParsedFields: partialParsed,
		ErrorOffset:  errorOffset,
		ErrorMessage: errMsg,
	}, err
}

func MarshalProtoBinary(msg *dynamicpb.Message) ([]byte, error) {
	return proto.Marshal(msg)
}

func checkTimestamps(msg *dynamicpb.Message, prefix string, warnings *[]Warning) {
	if msg == nil {
		return
	}

	md := msg.ProtoReflect().Descriptor()
	fullName := string(md.FullName())

	if fullName == "google.protobuf.Timestamp" {
		secondsField := md.Fields().ByName("seconds")
		nanosField := md.Fields().ByName("nanos")

		var seconds int64
		var nanos int32

		if secondsField != nil && msg.ProtoReflect().Has(secondsField) {
			seconds = msg.ProtoReflect().Get(secondsField).Int()
		}
		if nanosField != nil && msg.ProtoReflect().Has(nanosField) {
			nanos = int32(msg.ProtoReflect().Get(nanosField).Int())
		}

		if seconds < 0 || nanos < 0 || nanos > 999999999 {
			fieldName := prefix
			if fieldName == "" {
				fieldName = "timestamp"
			}
			*warnings = append(*warnings, Warning{
				Field:   fieldName,
				Message: fmt.Sprintf("invalid timestamp: seconds=%d, nanos=%d", seconds, nanos),
			})
		}
		return
	}

	msg.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		fieldName := string(fd.Name())
		if prefix != "" {
			fieldName = prefix + "." + fieldName
		}

		if fd.IsList() {
			list := v.List()
			for i := 0; i < list.Len(); i++ {
				elem := list.Get(i)
				if fd.Kind() == protoreflect.MessageKind {
					checkTimestamps(elem.Message().Interface().(*dynamicpb.Message),
						fmt.Sprintf("%s[%d]", fieldName, i), warnings)
				}
			}
		} else if fd.IsMap() {
			mp := v.Map()
			mp.Range(func(k protoreflect.MapKey, mv protoreflect.Value) bool {
				if fd.MapValue().Kind() == protoreflect.MessageKind {
					checkTimestamps(mv.Message().Interface().(*dynamicpb.Message),
						fmt.Sprintf("%s[%v]", fieldName, k.Interface()), warnings)
				}
				return true
			})
		} else if fd.Kind() == protoreflect.MessageKind {
			checkTimestamps(v.Message().Interface().(*dynamicpb.Message), fieldName, warnings)
		}

		return true
	})
}

func ValidateRoundTrip(original []byte, msg *dynamicpb.Message) error {
	reMarshaled, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to re-marshal: %w", err)
	}

	if !bytes.Equal(original, reMarshaled) {
		return errors.New("round-trip validation failed: binary data differs")
	}

	return nil
}
