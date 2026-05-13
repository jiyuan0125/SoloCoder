package parser

import (
	"bytes"
	"encoding/asn1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"
)

const (
	TagBoolean     = 1
	TagInteger     = 2
	TagBitString   = 3
	TagOctetString = 4
	TagNull        = 5
	TagOID         = 6
	TagUTF8String  = 12
	TagSequence    = 16
	TagSet         = 17
	TagUTCTime     = 23
	TagGeneralizedTime = 24
)

type TagClass int

const (
	ClassUniversal       TagClass = 0
	ClassApplication     TagClass = 1
	ClassContextSpecific TagClass = 2
	ClassPrivate         TagClass = 3
)

type ParseOptions struct {
	MaxDepth int
}

type ASN1Node struct {
	TagNumber   int
	TagClass    TagClass
	IsConstructed bool
	Length      int
	HeaderLength int
	DataOffset int
	DataLength int
	RawBytes    []byte
	Value       interface{}
	ValueString string
	Children    []*ASN1Node
}

type ParseResult struct {
	Root    *ASN1Node
	RawData []byte
	Warnings []string
}

func ParseASN1(data []byte, options ParseOptions) (*ParseResult, error) {
	if options.MaxDepth == 0 {
		options.MaxDepth = 100
	}

	result := &ParseResult{
		RawData:  data,
		Warnings: []string{},
	}

	root, err := parseNode(data, 0, 0, options.MaxDepth, result)
	if err != nil {
		result.Root = root
		return result, err
	}

	result.Root = root
	return result, nil
}

func parseNode(data []byte, offset, depth, maxDepth int, result *ParseResult) (*ASN1Node, error) {
	if depth > maxDepth {
		result.Warnings = append(result.Warnings, fmt.Sprintf("警告: 嵌套层级超过 %d 层，在偏移 %d 处", maxDepth, offset))
	}

	if offset >= len(data) {
		return nil, fmt.Errorf("数据不足，在偏移 %d 处", offset)
	}

	node := &ASN1Node{}
	currentOffset := offset

	tagByte := data[currentOffset]
	currentOffset++

	node.TagClass = TagClass((tagByte & 0xC0) >> 6)
	node.IsConstructed = (tagByte & 0x20) != 0
	node.TagNumber = int(tagByte & 0x1F)

	if node.TagNumber == 0x1F {
		tagValue := 0
		for currentOffset < len(data) {
			b := data[currentOffset]
			currentOffset++
			tagValue = (tagValue << 7) | (int(b) & 0x7F)
			if (b & 0x80) == 0 {
				break
			}
		}
		node.TagNumber = tagValue
	}

	if currentOffset >= len(data) {
		return node, fmt.Errorf("无法读取长度，在偏移 %d 处", currentOffset)
	}

	lengthByte := data[currentOffset]
	currentOffset++

	if lengthByte < 0x80 {
		node.Length = int(lengthByte)
	} else if lengthByte == 0x80 {
		start := currentOffset
		for currentOffset+1 < len(data) {
			if data[currentOffset] == 0x00 && data[currentOffset+1] == 0x00 {
				node.Length = currentOffset - start
				currentOffset += 2
				break
			}
			currentOffset++
		}
	} else {
		lengthBytesCount := int(lengthByte & 0x7F)
		if currentOffset+lengthBytesCount > len(data) {
			return node, fmt.Errorf("长度字段数据不足，在偏移 %d 处", currentOffset)
		}
		node.Length = 0
		for i := 0; i < lengthBytesCount; i++ {
			node.Length = (node.Length << 8) | int(data[currentOffset+i])
		}
		currentOffset += lengthBytesCount
	}

	node.HeaderLength = currentOffset - offset
	node.DataOffset = currentOffset
	node.DataLength = node.Length

	if node.DataOffset+node.DataLength > len(data) {
		return node, fmt.Errorf("值字段数据不足，在偏移 %d 处", currentOffset)
	}

	node.RawBytes = data[offset : currentOffset+node.DataLength]

	err := parseValue(node, data, currentOffset, depth, maxDepth, result)
	return node, err
}

func parseValue(node *ASN1Node, data []byte, valueOffset, depth, maxDepth int, result *ParseResult) error {
	valueData := data[valueOffset : valueOffset+node.DataLength]

	if node.TagClass == ClassUniversal {
		switch node.TagNumber {
		case TagBoolean:
			if len(valueData) != 1 {
				return fmt.Errorf("BOOLEAN值长度应为1字节，在偏移 %d 处", valueOffset)
			}
			node.Value = valueData[0] != 0x00
			node.ValueString = fmt.Sprintf("%v", node.Value)

		case TagInteger:
			var intValue big.Int
			intValue.SetBytes(valueData)
			node.Value = intValue.String()
			node.ValueString = intValue.String()

		case TagBitString:
			if len(valueData) < 1 {
				return fmt.Errorf("BIT STRING长度不足，在偏移 %d 处", valueOffset)
			}
			unusedBits := int(valueData[0])
			bitData := valueData[1:]
			node.Value = map[string]interface{}{
				"unused_bits": unusedBits,
				"bytes":       bitData,
			}
			node.ValueString = fmt.Sprintf("%d bits unused, %s", unusedBits, hex.EncodeToString(bitData))

		case TagOctetString:
			node.Value = valueData
			node.ValueString = fmt.Sprintf("0x%s", hex.EncodeToString(valueData))
			if isPrintable(valueData) {
				node.ValueString += fmt.Sprintf(" (%s)", string(valueData))
			}

		case TagNull:
			node.Value = nil
			node.ValueString = "NULL"

		case TagOID:
			oid, err := parseOID(valueData)
			if err != nil {
				return err
			}
			node.Value = oid
			node.ValueString = oidToString(oid)

		case TagUTF8String:
			node.Value = string(valueData)
			node.ValueString = fmt.Sprintf("%q", string(valueData))

		case TagSequence, TagSet:
			node.Children = []*ASN1Node{}
			childOffset := valueOffset
			endOffset := valueOffset + node.DataLength

			for childOffset < endOffset {
				child, err := parseNode(data, childOffset, depth+1, maxDepth, result)
				if err != nil {
					return err
				}
				node.Children = append(node.Children, child)
				childOffset += child.HeaderLength + child.DataLength
			}
			node.ValueString = fmt.Sprintf("(%d children)", len(node.Children))

		case TagUTCTime:
			t, err := parseUTCTime(valueData)
			if err != nil {
				return err
			}
			node.Value = t
			node.ValueString = t.Format(time.RFC3339)

		case TagGeneralizedTime:
			t, err := parseGeneralizedTime(valueData)
			if err != nil {
				return err
			}
			node.Value = t
			node.ValueString = t.Format(time.RFC3339)

		default:
			node.Value = valueData
			node.ValueString = fmt.Sprintf("0x%s", hex.EncodeToString(valueData))
		}
	} else {
		if node.IsConstructed {
			node.Children = []*ASN1Node{}
			childOffset := valueOffset
			endOffset := valueOffset + node.DataLength

			for childOffset < endOffset {
				child, err := parseNode(data, childOffset, depth+1, maxDepth, result)
				if err != nil {
					return err
				}
				node.Children = append(node.Children, child)
				childOffset += child.HeaderLength + child.DataLength
			}
			node.ValueString = fmt.Sprintf("(%d children)", len(node.Children))
		} else {
			node.Value = valueData
			node.ValueString = fmt.Sprintf("0x%s", hex.EncodeToString(valueData))
		}
	}

	return nil
}

func parseOID(data []byte) ([]int, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("空OID数据")
	}

	oid := []int{}
	firstByte := int(data[0])
	oid = append(oid, firstByte/40)
	oid = append(oid, firstByte%40)

	for i := 1; i < len(data); {
		value := 0
		for i < len(data) {
			b := int(data[i])
			i++
			value = (value << 7) | (b & 0x7F)
			if (b & 0x80) == 0 {
				break
			}
		}
		oid = append(oid, value)
	}

	return oid, nil
}

func oidToString(oid []int) string {
	parts := make([]string, len(oid))
	for i, v := range oid {
		parts[i] = fmt.Sprintf("%d", v)
	}
	return strings.Join(parts, ".")
}

func parseUTCTime(data []byte) (time.Time, error) {
	s := string(data)
	formats := []string{
		"060102150405Z",
		"0601021504Z",
		"060102150405-0700",
		"0601021504-0700",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			if t.Year() < 50 {
				t = t.AddDate(2000, 0, 0)
			} else {
				t = t.AddDate(1900, 0, 0)
			}
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("无法解析UTCTime: %s", s)
}

func parseGeneralizedTime(data []byte) (time.Time, error) {
	s := string(data)
	formats := []string{
		"20060102150405Z",
		"20060102150405",
		"200601021504Z",
		"200601021504",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("无法解析GeneralizedTime: %s", s)
}

func isPrintable(data []byte) bool {
	for _, b := range data {
		if b < 32 || b > 126 {
			return false
		}
	}
	return true
}

func getTagName(tagNumber int, tagClass TagClass) string {
	if tagClass == ClassUniversal {
		switch tagNumber {
		case TagBoolean:
			return "BOOLEAN"
		case TagInteger:
			return "INTEGER"
		case TagBitString:
			return "BIT STRING"
		case TagOctetString:
			return "OCTET STRING"
		case TagNull:
			return "NULL"
		case TagOID:
			return "OBJECT IDENTIFIER"
		case TagUTF8String:
			return "UTF8String"
		case TagSequence:
			return "SEQUENCE"
		case TagSet:
			return "SET"
		case TagUTCTime:
			return "UTCTime"
		case TagGeneralizedTime:
			return "GeneralizedTime"
		default:
			return fmt.Sprintf("UNKNOWN(%d)", tagNumber)
		}
	}

	className := ""
	switch tagClass {
	case ClassApplication:
		className = "Application"
	case ClassContextSpecific:
		className = "Context"
	case ClassPrivate:
		className = "Private"
	}
	return fmt.Sprintf("[%d] %s", tagNumber, className)
}

func (r *ParseResult) ToTreeString() string {
	if r.Root == nil {
		return ""
	}

	var buf bytes.Buffer
	writeNodeTree(&buf, r.Root, "", true)
	return buf.String()
}

func writeNodeTree(buf *bytes.Buffer, node *ASN1Node, prefix string, isLast bool) {
	tagName := getTagName(node.TagNumber, node.TagClass)
	if node.IsConstructed {
		buf.WriteString(fmt.Sprintf("%s%s%s (length=%d)\n", prefix, getConnector(isLast), tagName, node.Length))
	} else {
		buf.WriteString(fmt.Sprintf("%s%s%s (length=%d): %s\n", prefix, getConnector(isLast), tagName, node.Length, node.ValueString))
	}

	if len(node.Children) > 0 {
		childPrefix := prefix + getChildPrefix(isLast)
		for i, child := range node.Children {
			writeNodeTree(buf, child, childPrefix, i == len(node.Children)-1)
		}
	}
}

func getConnector(isLast bool) string {
	if isLast {
		return "└── "
	}
	return "├── "
}

func getChildPrefix(isLast bool) string {
	if isLast {
		return "    "
	}
	return "│   "
}

func (r *ParseResult) ToJSON() (string, error) {
	if r.Root == nil {
		return "null", nil
	}

	jsonData := jsonNode(r.Root)
	jsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func jsonNode(node *ASN1Node) map[string]interface{} {
	result := map[string]interface{}{
		"tag_number":   node.TagNumber,
		"tag_class":    node.TagClass,
		"is_constructed": node.IsConstructed,
		"tag_name":     getTagName(node.TagNumber, node.TagClass),
		"length":       node.Length,
		"raw_bytes":    hex.EncodeToString(node.RawBytes),
	}

	if node.TagClass == ClassUniversal {
		result["value"] = node.Value
	} else {
		result["value_hex"] = hex.EncodeToString(node.RawBytes[node.HeaderLength:])
	}

	if len(node.Children) > 0 {
		children := make([]map[string]interface{}, len(node.Children))
		for i, child := range node.Children {
			children[i] = jsonNode(child)
		}
		result["children"] = children
	}

	return result
}

func (r *ParseResult) ToDER() ([]byte, error) {
	if r.Root == nil {
		return nil, fmt.Errorf("没有可编码的数据")
	}
	return encodeNode(r.Root)
}

func encodeNode(node *ASN1Node) ([]byte, error) {
	var buf bytes.Buffer

	if len(node.Children) > 0 {
		var childrenData bytes.Buffer
		for _, child := range node.Children {
			childBytes, err := encodeNode(child)
			if err != nil {
				return nil, err
			}
			childrenData.Write(childBytes)
		}
		encoded, err := asn1.Marshal(childrenData.Bytes())
		if err != nil {
			return nil, err
		}
		buf.Write(encoded)
	} else {
		var data []byte
		var err error

		switch node.TagNumber {
		case TagInteger:
			var intVal big.Int
			intVal.SetString(node.Value.(string), 10)
			data, err = asn1.Marshal(intVal)
		case TagNull:
			data, err = asn1.Marshal(asn1.NullRawValue)
		case TagOID:
			oid := asn1.ObjectIdentifier(node.Value.([]int))
			data, err = asn1.Marshal(oid)
		case TagUTF8String:
			data, err = asn1.MarshalWithParams(node.Value.(string), "utf8")
		case TagOctetString:
			data, err = asn1.Marshal(node.Value.([]byte))
		default:
			data = node.RawBytes
			err = nil
		}

		if err != nil {
			return nil, err
		}
		buf.Write(data)
	}

	return buf.Bytes(), nil
}
