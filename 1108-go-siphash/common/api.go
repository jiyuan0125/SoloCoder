package common

type Variant string

const (
	VariantSipHash24 Variant = "siphash-2-4"
	VariantSipHash13 Variant = "siphash-1-3"
)

type HashRequest struct {
	Data    []byte  `json:"data"`
	Variant Variant `json:"variant,omitempty"`
}

type HashResponse struct {
	Hash   uint64  `json:"hash"`
	Hex    string  `json:"hex"`
	Variant Variant `json:"variant"`
	Error  string  `json:"error,omitempty"`
}

type PutRequest struct {
	Key   []byte `json:"key"`
	Value []byte `json:"value"`
}

type PutResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

type GetRequest struct {
	Key []byte `json:"key"`
}

type GetResponse struct {
	Value []byte `json:"value"`
	Found bool   `json:"found"`
	Error string `json:"error,omitempty"`
}

type DeleteRequest struct {
	Key []byte `json:"key"`
}

type DeleteResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

type BatchPutItem struct {
	Key   []byte `json:"key"`
	Value []byte `json:"value"`
}

type BatchPutRequest struct {
	Items []BatchPutItem `json:"items"`
}

type BatchPutResponse struct {
	Success int    `json:"success"`
	Error   string `json:"error,omitempty"`
}

type BatchGetRequest struct {
	Keys [][]byte `json:"keys"`
}

type BatchGetItem struct {
	Key   []byte `json:"key"`
	Value []byte `json:"value"`
	Found bool   `json:"found"`
}

type BatchGetResponse struct {
	Items []BatchGetItem `json:"items"`
	Error string         `json:"error,omitempty"`
}

type BatchDeleteRequest struct {
	Keys [][]byte `json:"keys"`
}

type BatchDeleteResponse struct {
	Deleted int    `json:"deleted"`
	Error   string `json:"error,omitempty"`
}

type RotateKeyRequest struct {
	NewKey []byte `json:"new_key"`
}

type RotateKeyResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

type ListResponse struct {
	Keys  [][]byte `json:"keys"`
	Count int      `json:"count"`
	Error string   `json:"error,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
