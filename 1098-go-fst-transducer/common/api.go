package common

type KeyValue struct {
	Key   string `json:"key"`
	Value uint64 `json:"value"`
}

type BuildRequest struct {
	DictName string     `json:"dict_name"`
	Items    []KeyValue `json:"items"`
}

type BuildResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Count   int    `json:"count,omitempty"`
}

type SearchRequest struct {
	DictName string `json:"dict_name"`
	Key      string `json:"key"`
}

type SearchResponse struct {
	Found bool   `json:"found"`
	Value uint64 `json:"value,omitempty"`
	Error string `json:"error,omitempty"`
}

type PrefixRequest struct {
	DictName string `json:"dict_name"`
	Prefix   string `json:"prefix"`
}

type PrefixResponse struct {
	Items []KeyValue `json:"items"`
	Error string     `json:"error,omitempty"`
}

type DumpResponse struct {
	Items []KeyValue `json:"items"`
	Error string     `json:"error,omitempty"`
}

type SaveRequest struct {
	DictName string `json:"dict_name"`
	FilePath string `json:"file_path"`
}

type LoadRequest struct {
	DictName string `json:"dict_name"`
	FilePath string `json:"file_path"`
}

type SaveLoadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ListResponse struct {
	Dicts []string `json:"dicts"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
