package api

type PutRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type PutResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type GetRequest struct {
	Key string `json:"key"`
}

type GetResponse struct {
	Value   string `json:"value"`
	Found   bool   `json:"found"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type DeleteRequest struct {
	Key string `json:"key"`
}

type DeleteResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type ScanRequest struct {
	Start string `json:"start,omitempty"`
	End   string `json:"end,omitempty"`
}

type KVPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ScanResponse struct {
	Pairs   []KVPair `json:"pairs"`
	Success bool     `json:"success"`
	Error   string   `json:"error,omitempty"`
}
