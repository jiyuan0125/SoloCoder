package common

import "encoding/json"

type ParseRequest struct {
	Data []byte `json:"data"`
}

type ParseResponse struct {
	Fields  []*Field `json:"fields"`
	Error   string   `json:"error,omitempty"`
	Success bool     `json:"success"`
}

type Field struct {
	FieldNumber int         `json:"field_number"`
	WireType    int         `json:"wire_type"`
	WireTypeName string     `json:"wire_type_name"`
	RawBytes    []byte      `json:"raw_bytes"`
	RawByteSize int         `json:"raw_byte_size"`
	Value       interface{} `json:"value"`
	ValueType   string      `json:"value_type"`
	Repeated    bool        `json:"repeated,omitempty"`
}

type NestedMessage struct {
	Fields []*Field `json:"fields"`
}

func WireTypeName(wt int) string {
	switch wt {
	case 0:
		return "varint"
	case 1:
		return "64-bit fixed"
	case 2:
		return "length-delimited"
	case 3:
		return "start group"
	case 4:
		return "end group"
	case 5:
		return "32-bit fixed"
	default:
		return "unknown"
	}
}

func (r *ParseResponse) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
