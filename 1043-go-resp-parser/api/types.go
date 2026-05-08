package api

type Request struct {
	Commands [][]string `json:"commands"`
}

type ResponseValue struct {
	Type  string `json:"type"`
	Str   string `json:"str,omitempty"`
	Int   *int64 `json:"int,omitempty"`
	Array []ResponseValue `json:"array,omitempty"`
	IsNull bool  `json:"is_null,omitempty"`
}

type Response struct {
	Results []ResponseValue `json:"results"`
	Error   string          `json:"error,omitempty"`
}

func NewResponseError(err string) Response {
	return Response{Error: err}
}

func ValueTypeToString(t int) string {
	switch t {
	case 0:
		return "simple_string"
	case 1:
		return "error"
	case 2:
		return "integer"
	case 3:
		return "bulk_string"
	case 4:
		return "array"
	case 5:
		return "null_bulk_string"
	case 6:
		return "null_array"
	default:
		return "unknown"
	}
}
