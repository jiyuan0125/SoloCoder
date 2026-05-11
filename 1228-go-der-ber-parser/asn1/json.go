package asn1

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

type JSONTLV struct {
	Class      string      `json:"class,omitempty"`
	Tag        string      `json:"tag,omitempty"`
	TagNumber  uint32      `json:"tag_number,omitempty"`
	Constructed bool       `json:"constructed,omitempty"`
	Type       string      `json:"type,omitempty"`
	Value      interface{} `json:"value,omitempty"`
	Children   []JSONTLV   `json:"children,omitempty"`
}

func tagNumberToString(num TagNumber, class TagClass) string {
	if class != TagClassUniversal {
		return fmt.Sprintf("[%d]", num)
	}
	switch num {
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
	case TagObjectIdentifier:
		return "OBJECT IDENTIFIER"
	case TagSequence:
		return "SEQUENCE"
	case TagSet:
		return "SET"
	case TagPrintableString:
		return "PrintableString"
	case TagT61String:
		return "T61String"
	case TagIA5String:
		return "IA5String"
	case TagUTCTime:
		return "UTCTime"
	case TagGeneralizedTime:
		return "GeneralizedTime"
	case TagUTF8String:
		return "UTF8String"
	default:
		return fmt.Sprintf("[%d]", num)
	}
}

func tagClassToString(class TagClass) string {
	switch class {
	case TagClassUniversal:
		return "universal"
	case TagClassApplication:
		return "application"
	case TagClassContextSpecific:
		return "context-specific"
	case TagClassPrivate:
		return "private"
	default:
		return "unknown"
	}
}

func parseTagClass(s string) (TagClass, error) {
	switch strings.ToLower(s) {
	case "universal":
		return TagClassUniversal, nil
	case "application":
		return TagClassApplication, nil
	case "context-specific":
		return TagClassContextSpecific, nil
	case "private":
		return TagClassPrivate, nil
	default:
		return TagClassUniversal, fmt.Errorf("unknown tag class: %s", s)
	}
}

func parseTagType(s string) (TagNumber, TagClass, bool, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		numStr := s[1 : len(s)-1]
		var num uint32
		_, err := fmt.Sscanf(numStr, "%d", &num)
		if err != nil {
			return 0, TagClassUniversal, false, err
		}
		return TagNumber(num), TagClassContextSpecific, true, nil
	}

	switch strings.ToUpper(s) {
	case "BOOLEAN":
		return TagBoolean, TagClassUniversal, false, nil
	case "INTEGER":
		return TagInteger, TagClassUniversal, false, nil
	case "BIT STRING", "BITSTRING":
		return TagBitString, TagClassUniversal, false, nil
	case "OCTET STRING", "OCTETSTRING":
		return TagOctetString, TagClassUniversal, false, nil
	case "NULL":
		return TagNull, TagClassUniversal, false, nil
	case "OBJECT IDENTIFIER", "OID":
		return TagObjectIdentifier, TagClassUniversal, false, nil
	case "SEQUENCE":
		return TagSequence, TagClassUniversal, true, nil
	case "SET":
		return TagSet, TagClassUniversal, true, nil
	case "PRINTABLESTRING":
		return TagPrintableString, TagClassUniversal, false, nil
	case "T61STRING":
		return TagT61String, TagClassUniversal, false, nil
	case "IA5STRING":
		return TagIA5String, TagClassUniversal, false, nil
	case "UTCTIME":
		return TagUTCTime, TagClassUniversal, false, nil
	case "GENERALIZEDTIME":
		return TagGeneralizedTime, TagClassUniversal, false, nil
	case "UTF8STRING":
		return TagUTF8String, TagClassUniversal, false, nil
	default:
		return 0, TagClassUniversal, false, fmt.Errorf("unknown tag type: %s", s)
	}
}

func (t *TLV) ToJSON() (JSONTLV, error) {
	j := JSONTLV{
		Class:       tagClassToString(t.Class),
		Tag:         tagNumberToString(t.Number, t.Class),
		TagNumber:   uint32(t.Number),
		Constructed: t.Constructed,
		Type:        tagNumberToString(t.Number, t.Class),
	}

	if t.Constructed {
		j.Children = make([]JSONTLV, 0, len(t.Children))
		for _, child := range t.Children {
			cj, err := child.ToJSON()
			if err != nil {
				return j, err
			}
			j.Children = append(j.Children, cj)
		}
	} else {
		j.Value = formatValue(t)
	}

	return j, nil
}

func formatValue(t *TLV) interface{} {
	switch t.Number {
	case TagBoolean:
		if len(t.Value) > 0 {
			return t.Value[0] != 0
		}
		return false
	case TagNull:
		return nil
	case TagInteger:
		if len(t.Value) == 0 {
			return 0
		}
		var val int64
		for _, b := range t.Value {
			val = (val << 8) | int64(b)
		}
		return val
	case TagOctetString, TagPrintableString, TagT61String, TagIA5String, TagUTF8String:
		return string(t.Value)
	case TagBitString:
		return hex.EncodeToString(t.Value)
	default:
		return hex.EncodeToString(t.Value)
	}
}

func parseValue(jt JSONTLV, num TagNumber) ([]byte, error) {
	if jt.Value == nil {
		return nil, nil
	}

	switch num {
	case TagBoolean:
		if v, ok := jt.Value.(bool); ok {
			if v {
				return []byte{0xFF}, nil
			}
			return []byte{0x00}, nil
		}
		return nil, fmt.Errorf("boolean value expected")

	case TagInteger:
		switch v := jt.Value.(type) {
		case float64:
			return encodeInteger(int64(v))
		case int:
			return encodeInteger(int64(v))
		case int64:
			return encodeInteger(v)
		case string:
			var val int64
			_, err := fmt.Sscanf(v, "%d", &val)
			if err != nil {
				return nil, err
			}
			return encodeInteger(val)
		}
		return nil, fmt.Errorf("integer value expected")

	case TagNull:
		return nil, nil

	case TagOctetString, TagPrintableString, TagT61String, TagIA5String, TagUTF8String:
		if v, ok := jt.Value.(string); ok {
			return []byte(v), nil
		}
		return nil, fmt.Errorf("string value expected")

	case TagBitString:
		if v, ok := jt.Value.(string); ok {
			return hex.DecodeString(v)
		}
		return nil, fmt.Errorf("hex string expected for bit string")

	default:
		if v, ok := jt.Value.(string); ok {
			return hex.DecodeString(v)
		}
		return nil, fmt.Errorf("hex string expected")
	}
}

func encodeInteger(val int64) ([]byte, error) {
	if val == 0 {
		return []byte{0x00}, nil
	}

	var bytes []byte
	n := val
	if n > 0 {
		for n > 0 {
			bytes = append([]byte{byte(n & 0xFF)}, bytes...)
			n >>= 8
		}
		if bytes[0]&0x80 != 0 {
			bytes = append([]byte{0x00}, bytes...)
		}
	} else {
		n = -n - 1
		for n > 0 {
			bytes = append([]byte{byte(0xFF - (n & 0xFF))}, bytes...)
			n >>= 8
		}
		if bytes[0]&0x80 == 0 {
			bytes = append([]byte{0xFF}, bytes...)
		}
	}

	return bytes, nil
}

func FromJSON(jt JSONTLV) (*TLV, error) {
	var num TagNumber
	var class TagClass
	var constructed bool
	var err error

	if jt.Type != "" {
		num, class, constructed, err = parseTagType(jt.Type)
		if err != nil {
			return nil, err
		}
	} else if jt.Tag != "" {
		num, class, constructed, err = parseTagType(jt.Tag)
		if err != nil {
			return nil, err
		}
	} else {
		num = TagNumber(jt.TagNumber)
		class = TagClassUniversal
		constructed = jt.Constructed
	}

	if jt.Class != "" {
		class, err = parseTagClass(jt.Class)
		if err != nil {
			return nil, err
		}
	}

	tlv := NewTLV(class, num, constructed)

	if constructed {
		for _, childJT := range jt.Children {
			child, err := FromJSON(childJT)
			if err != nil {
				return nil, err
			}
			tlv.AddChild(child)
		}
	} else {
		value, err := parseValue(jt, num)
		if err != nil {
			return nil, err
		}
		tlv.Value = value
	}

	return tlv, nil
}

func (j JSONTLV) Marshal() ([]byte, error) {
	return json.MarshalIndent(j, "", "  ")
}

func UnmarshalJSONTLV(data []byte) (*JSONTLV, error) {
	var jt JSONTLV
	err := json.Unmarshal(data, &jt)
	if err != nil {
		return nil, err
	}
	return &jt, nil
}
