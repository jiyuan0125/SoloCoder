package api

type SortRequest struct {
	Data []int64 `json:"data"`
}

type SortResponse struct {
	Success bool    `json:"success"`
	Data    []int64 `json:"data,omitempty"`
	Error   string  `json:"error,omitempty"`
	Stats   Stats   `json:"stats"`
}

type RadixRequest struct {
	Radix int `json:"radix"`
}

type RadixResponse struct {
	Success bool   `json:"success"`
	Radix   int    `json:"radix,omitempty"`
	Error   string `json:"error,omitempty"`
}

type StatsResponse struct {
	Success bool   `json:"success"`
	Stats   Stats  `json:"stats"`
	Error   string `json:"error,omitempty"`
}

type Stats struct {
	Passes          int `json:"passes"`
	TotalOperations int `json:"total_operations"`
	ArraySize       int `json:"array_size"`
	Radix           int `json:"radix"`
}

type StructField struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

type StructSortRequest struct {
	Data      []StructField `json:"data"`
	FieldName string        `json:"field_name"`
}

type StructSortResponse struct {
	Success bool          `json:"success"`
	Data    []StructField `json:"data,omitempty"`
	Error   string        `json:"error,omitempty"`
	Stats   Stats         `json:"stats"`
}
