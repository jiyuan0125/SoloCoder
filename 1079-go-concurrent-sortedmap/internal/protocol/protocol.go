package protocol

import "encoding/json"

type KVPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

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
	Success bool   `json:"success"`
	Value   string `json:"value,omitempty"`
	Exists  bool   `json:"exists"`
	Error   string `json:"error,omitempty"`
}

type DeleteRequest struct {
	Key string `json:"key"`
}

type DeleteResponse struct {
	Success bool `json:"success"`
	Deleted bool `json:"deleted"`
	Error   string `json:"error,omitempty"`
}

type RangeRequest struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type RangeResponse struct {
	Success bool     `json:"success"`
	Pairs   []KVPair `json:"pairs,omitempty"`
	Error   string   `json:"error,omitempty"`
}

type SizeResponse struct {
	Success bool  `json:"success"`
	Size    int64 `json:"size,omitempty"`
	Error   string `json:"error,omitempty"`
}

func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
