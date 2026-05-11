package common

type FieldInfo struct {
	FlatKey         string   `json:"flat_key"`
	OriginalPath    string   `json:"original_path"`
	Type            string   `json:"type"`
	IsCircular      bool     `json:"is_circular,omitempty"`
	HasMultiplePaths bool    `json:"has_multiple_paths,omitempty"`
	IsArrayIndex    bool     `json:"is_array_index,omitempty"`
	IsMapKey        bool     `json:"is_map_key,omitempty"`
	HasOmitempty    bool     `json:"has_omitempty,omitempty"`
	Notes           []string `json:"notes,omitempty"`
}

type FlattenRequest struct {
	SourceCode string `json:"source_code"`
	StructName string `json:"struct_name"`
}

type FlattenResponse struct {
	Success bool         `json:"success"`
	Fields  []FieldInfo  `json:"fields,omitempty"`
	Error   string       `json:"error,omitempty"`
}

type ListStructsRequest struct {
	SourceCode string `json:"source_code"`
}

type ListStructsResponse struct {
	Success  bool     `json:"success"`
	Structs  []string `json:"structs,omitempty"`
	Error    string   `json:"error,omitempty"`
}
