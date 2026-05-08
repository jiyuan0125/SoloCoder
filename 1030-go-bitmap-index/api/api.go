package api

type ErrorResponse struct {
	Error string `json:"error"`
}

type CreateIndexRequest struct {
	Name string `json:"name"`
	Size uint64 `json:"size"`
}

type SetBitRequest struct {
	Index string `json:"index"`
	Bit   uint64 `json:"bit"`
}

type ClearBitRequest struct {
	Index string `json:"index"`
	Bit   uint64 `json:"bit"`
}

type GetBitRequest struct {
	Index string `json:"index"`
	Bit   uint64 `json:"bit"`
}

type GetBitResponse struct {
	Value bool `json:"value"`
}

type CountRequest struct {
	Index string `json:"index"`
}

type CountResponse struct {
	Count uint64 `json:"count"`
}

type IndicesRequest struct {
	Index string `json:"index"`
}

type IndicesResponse struct {
	Indices []uint64 `json:"indices"`
}

type Operation string

const (
	OpAnd    Operation = "and"
	OpOr     Operation = "or"
	OpXor    Operation = "xor"
	OpNot    Operation = "not"
	OpAndNot Operation = "andnot"
)

type OperationRequest struct {
	Operation Operation `json:"operation"`
	Indices   []string  `json:"indices"`
	Result    string    `json:"result"`
}

type OperationResponse struct {
	Count uint64 `json:"count"`
}

type ListIndicesResponse struct {
	Indices []IndexInfo `json:"indices"`
}

type IndexInfo struct {
	Name  string `json:"name"`
	Size  uint64 `json:"size"`
	Count uint64 `json:"count"`
}

type DeleteIndexRequest struct {
	Name string `json:"name"`
}

type SerializeRequest struct {
	Index string `json:"index"`
}

type SerializeResponse struct {
	Data []byte `json:"data"`
}

type DeserializeRequest struct {
	Name string `json:"name"`
	Data []byte `json:"data"`
}
