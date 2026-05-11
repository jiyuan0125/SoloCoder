package api

type PutRequest struct {
	Key   int64 `json:"key"`
	Value int64 `json:"value"`
}

type PutResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type GetResponse struct {
	Success bool   `json:"success"`
	Key     int64  `json:"key,omitempty"`
	Value   int64  `json:"value,omitempty"`
	Error   string `json:"error,omitempty"`
}

type DeleteResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type PrevNextResponse struct {
	Success bool   `json:"success"`
	Key     int64  `json:"key,omitempty"`
	Value   int64  `json:"value,omitempty"`
	Exists  bool   `json:"exists"`
	Error   string `json:"error,omitempty"`
}

type ListResponse struct {
	Success bool          `json:"success"`
	Items   []KeyValuePair `json:"items,omitempty"`
	Error   string        `json:"error,omitempty"`
}

type KeyValuePair struct {
	Key   int64 `json:"key"`
	Value int64 `json:"value"`
}
