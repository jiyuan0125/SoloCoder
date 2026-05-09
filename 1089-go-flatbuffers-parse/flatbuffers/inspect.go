package flatbuffers

import (
	"fmt"
	"strings"
)

type FieldSchema struct {
	Name     string
	Type     FieldType
	ElemType FieldType
	Nested   *TableSchema
}

type TableSchema struct {
	Name   string
	Fields []FieldSchema
}

type InspectNode struct {
	Name     string
	Value    interface{}
	Type     string
	Children []InspectNode
}

func Inspect(buf []byte, schema *TableSchema) (*InspectNode, error) {
	tablePos, err := GetRootAsTable(buf)
	if err != nil {
		return nil, err
	}
	return inspectTable(buf, tablePos, schema, schema.Name)
}

func inspectTable(buf []byte, tablePos int, schema *TableSchema, name string) (*InspectNode, error) {
	node := &InspectNode{
		Name:     name,
		Type:     "table",
		Children: make([]InspectNode, 0),
	}

	for i, field := range schema.Fields {
		fieldNode, err := inspectField(buf, tablePos, i, &field)
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, *fieldNode)
	}

	return node, nil
}

func inspectField(buf []byte, tablePos int, fieldIndex int, field *FieldSchema) (*InspectNode, error) {
	exists, err := HasField(buf, tablePos, fieldIndex)
	if err != nil {
		return nil, err
	}

	node := &InspectNode{
		Name: field.Name,
		Type: fieldTypeName(field.Type),
	}

	if !exists {
		node.Value = getDefaultValue(field.Type)
		return node, nil
	}

	value, err := GetField(buf, tablePos, fieldIndex, field.Type)
	if err != nil {
		return nil, err
	}

	switch field.Type {
	case TypeTable:
		nestedPos := value.(int)
		if field.Nested != nil {
			nestedNode, err := inspectTable(buf, nestedPos, field.Nested, field.Name)
			if err != nil {
				return nil, err
			}
			return nestedNode, nil
		} else {
			node.Value = fmt.Sprintf("table[offset=%d]", nestedPos)
		}
	case TypeScalarArray:
		arrayInfo := value.(*ArrayInfo)
		node.Type = fmt.Sprintf("[%s]", fieldTypeName(field.ElemType))
		node.Children = make([]InspectNode, 0, arrayInfo.Length)
		for i := 0; i < arrayInfo.Length; i++ {
			elemValue, err := ReadScalarArrayElement(buf, arrayInfo.Pos, i, field.ElemType)
			if err != nil {
				return nil, err
			}
			elemNode := InspectNode{
				Name:  fmt.Sprintf("[%d]", i),
				Type:  fieldTypeName(field.ElemType),
				Value: elemValue,
			}
			node.Children = append(node.Children, elemNode)
		}
	case TypeTableArray:
		arrayInfo := value.(*ArrayInfo)
		node.Type = fmt.Sprintf("[table]")
		node.Children = make([]InspectNode, 0, arrayInfo.Length)
		for i := 0; i < arrayInfo.Length; i++ {
			elemPos, err := ReadTableArrayElement(buf, arrayInfo.Pos, i)
			if err != nil {
				return nil, err
			}
			if field.Nested != nil {
				elemNode, err := inspectTable(buf, elemPos, field.Nested, fmt.Sprintf("[%d]", i))
				if err != nil {
					return nil, err
				}
				node.Children = append(node.Children, *elemNode)
			} else {
				elemNode := InspectNode{
					Name:  fmt.Sprintf("[%d]", i),
					Type:  "table",
					Value: fmt.Sprintf("table[offset=%d]", elemPos),
				}
				node.Children = append(node.Children, elemNode)
			}
		}
	default:
		node.Value = value
	}

	return node, nil
}

func fieldTypeName(t FieldType) string {
	switch t {
	case TypeInt8:
		return "int8"
	case TypeUint8:
		return "uint8"
	case TypeInt16:
		return "int16"
	case TypeUint16:
		return "uint16"
	case TypeInt32:
		return "int32"
	case TypeUint32:
		return "uint32"
	case TypeInt64:
		return "int64"
	case TypeUint64:
		return "uint64"
	case TypeFloat32:
		return "float32"
	case TypeFloat64:
		return "float64"
	case TypeBool:
		return "bool"
	case TypeString:
		return "string"
	case TypeTable:
		return "table"
	case TypeScalarArray:
		return "scalar_array"
	case TypeTableArray:
		return "table_array"
	default:
		return "unknown"
	}
}

func (n *InspectNode) String() string {
	return n.stringIndent(0)
}

func (n *InspectNode) stringIndent(indent int) string {
	var sb strings.Builder
	prefix := strings.Repeat("  ", indent)
	sb.WriteString(prefix)
	sb.WriteString(n.Name)
	sb.WriteString(" (")
	sb.WriteString(n.Type)
	sb.WriteString(")")
	if n.Value != nil {
		sb.WriteString(": ")
		sb.WriteString(fmt.Sprintf("%v", n.Value))
	}
	sb.WriteString("\n")
	for _, child := range n.Children {
		sb.WriteString(child.stringIndent(indent + 1))
	}
	return sb.String()
}

func (n *InspectNode) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	result["name"] = n.Name
	result["type"] = n.Type
	if n.Value != nil {
		result["value"] = n.Value
	}
	if len(n.Children) > 0 {
		children := make([]map[string]interface{}, len(n.Children))
		for i, child := range n.Children {
			children[i] = child.ToMap()
		}
		result["children"] = children
	}
	return result
}
