package schema

import (
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

type SimpleProtoParser struct{}

func NewSimpleProtoParser() *SimpleProtoParser {
	return &SimpleProtoParser{}
}

func (p *SimpleProtoParser) Parse(protoContent string, messageName string) (protoreflect.MessageDescriptor, error) {
	desc, err := p.parseProtoToDescriptor(protoContent, messageName)
	if err != nil {
		return nil, err
	}
	return desc, nil
}

func (p *SimpleProtoParser) parseProtoToDescriptor(protoContent string, targetMessageName string) (protoreflect.MessageDescriptor, error) {
	lines := strings.Split(protoContent, "\n")
	var packageName string
	messages := make(map[string]*descriptorpb.DescriptorProto)
	enums := make(map[string]*descriptorpb.EnumDescriptorProto)
	currentMessage := ""
	messageStack := []string{}
	syntax := "proto3"

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		if strings.HasPrefix(line, "syntax") {
			if idx := strings.Index(line, `"`); idx != -1 {
				endIdx := strings.Index(line[idx+1:], `"`)
				if endIdx != -1 {
					syntax = line[idx+1 : idx+1+endIdx]
				}
			}
			continue
		}
		if strings.HasPrefix(line, "package") {
			packageName = strings.TrimSuffix(strings.TrimPrefix(line, "package "), ";")
			continue
		}
		if strings.HasPrefix(line, "message ") {
			msgName := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "message "), "{"))
			msgName = strings.TrimSpace(msgName)
			if currentMessage != "" {
				messageStack = append(messageStack, currentMessage)
			}
			currentMessage = msgName
			messages[currentMessage] = &descriptorpb.DescriptorProto{
				Name: protoString(currentMessage),
			}
			continue
		}
		if strings.HasPrefix(line, "enum ") {
			enumName := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "enum "), "{"))
			enumName = strings.TrimSpace(enumName)
			enums[enumName] = &descriptorpb.EnumDescriptorProto{
				Name: protoString(enumName),
			}
			continue
		}
		if line == "}" {
			if len(messageStack) > 0 {
				parent := messageStack[len(messageStack)-1]
				messageStack = messageStack[:len(messageStack)-1]
				if msg, ok := messages[currentMessage]; ok {
					if parentMsg, ok := messages[parent]; ok {
						parentMsg.NestedType = append(parentMsg.NestedType, msg)
					}
				}
				currentMessage = parent
			} else {
				currentMessage = ""
			}
			continue
		}
		if currentMessage != "" && strings.Contains(line, "=") {
			msg := messages[currentMessage]
			field, err := p.parseField(line, syntax)
			if err == nil && msg != nil {
				msg.Field = append(msg.Field, field)
			}
		}
	}

	targetMsg, ok := messages[targetMessageName]
	if !ok {
		return nil, fmt.Errorf("message %s not found in proto definition", targetMessageName)
	}

	fileDescProto := &descriptorpb.FileDescriptorProto{
		Name:   protoString("dynamic.proto"),
		Package: protoString(packageName),
		Syntax: protoString(syntax),
		MessageType: []*descriptorpb.DescriptorProto{targetMsg},
	}

	fileDesc, err := protodesc.NewFile(fileDescProto, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create file descriptor: %w", err)
	}

	msgDesc := fileDesc.Messages().ByName(protoreflect.Name(targetMessageName))
	if msgDesc == nil {
		return nil, fmt.Errorf("message descriptor not found")
	}

	return msgDesc, nil
}

func (p *SimpleProtoParser) parseField(line string, syntax string) (*descriptorpb.FieldDescriptorProto, error) {
	eqIdx := strings.Index(line, "=")
	if eqIdx == -1 {
		return nil, fmt.Errorf("invalid field line (no '='): %s", line)
	}

	beforeEq := strings.TrimSpace(line[:eqIdx])
	afterEq := strings.TrimSpace(line[eqIdx+1:])

	semicolonIdx := strings.Index(afterEq, ";")
	if semicolonIdx == -1 {
		return nil, fmt.Errorf("invalid field line (no ';'): %s", line)
	}
	numStr := strings.TrimSpace(afterEq[:semicolonIdx])
	n, err := strconv.ParseInt(numStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid field number %q: %w", numStr, err)
	}
	fieldNumber := int32(n)

	beforeParts := strings.Fields(beforeEq)
	if len(beforeParts) < 2 {
		return nil, fmt.Errorf("invalid field definition before '=': %s", beforeEq)
	}

	var label descriptorpb.FieldDescriptorProto_Label
	label = descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	idx := 0

	if syntax == "proto2" {
		if beforeParts[idx] == "required" {
			label = descriptorpb.FieldDescriptorProto_LABEL_REQUIRED
			idx++
		} else if beforeParts[idx] == "optional" {
			label = descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
			idx++
		}
	}

	if idx < len(beforeParts) && beforeParts[idx] == "repeated" {
		label = descriptorpb.FieldDescriptorProto_LABEL_REPEATED
		idx++
	}

	if idx+2 > len(beforeParts) {
		return nil, fmt.Errorf("invalid field definition: %s", beforeEq)
	}

	fieldType := beforeParts[idx]
	fieldName := beforeParts[idx+1]

	kind := typeToKind(fieldType)
	var typeName *string
	if kind == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE || kind == descriptorpb.FieldDescriptorProto_TYPE_ENUM {
		typeName = protoString("." + fieldType)
	}

	labelCopy := label
	kindCopy := kind
	fieldNumberCopy := fieldNumber

	return &descriptorpb.FieldDescriptorProto{
		Name:     protoString(fieldName),
		Number:   &fieldNumberCopy,
		Label:    &labelCopy,
		Type:     &kindCopy,
		TypeName: typeName,
	}, nil
}

func typeToKind(t string) descriptorpb.FieldDescriptorProto_Type {
	switch t {
	case "double":
		return descriptorpb.FieldDescriptorProto_TYPE_DOUBLE
	case "float":
		return descriptorpb.FieldDescriptorProto_TYPE_FLOAT
	case "int32":
		return descriptorpb.FieldDescriptorProto_TYPE_INT32
	case "int64":
		return descriptorpb.FieldDescriptorProto_TYPE_INT64
	case "uint32":
		return descriptorpb.FieldDescriptorProto_TYPE_UINT32
	case "uint64":
		return descriptorpb.FieldDescriptorProto_TYPE_UINT64
	case "sint32":
		return descriptorpb.FieldDescriptorProto_TYPE_SINT32
	case "sint64":
		return descriptorpb.FieldDescriptorProto_TYPE_SINT64
	case "fixed32":
		return descriptorpb.FieldDescriptorProto_TYPE_FIXED32
	case "fixed64":
		return descriptorpb.FieldDescriptorProto_TYPE_FIXED64
	case "sfixed32":
		return descriptorpb.FieldDescriptorProto_TYPE_SFIXED32
	case "sfixed64":
		return descriptorpb.FieldDescriptorProto_TYPE_SFIXED64
	case "bool":
		return descriptorpb.FieldDescriptorProto_TYPE_BOOL
	case "string":
		return descriptorpb.FieldDescriptorProto_TYPE_STRING
	case "bytes":
		return descriptorpb.FieldDescriptorProto_TYPE_BYTES
	default:
		return descriptorpb.FieldDescriptorProto_TYPE_MESSAGE
	}
}

func protoString(s string) *string {
	return &s
}

func protoInt32(i int32) *int32 {
	return &i
}

func CreateDynamicMessage(desc protoreflect.MessageDescriptor) *dynamicpb.Message {
	return dynamicpb.NewMessage(desc)
}
