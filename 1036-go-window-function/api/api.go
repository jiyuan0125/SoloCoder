package api

import "encoding/json"

type Record map[string]interface{}

type WindowFuncType string

const (
	FuncRowNumber WindowFuncType = "ROW_NUMBER"
	FuncRank      WindowFuncType = "RANK"
	FuncDenseRank WindowFuncType = "DENSE_RANK"
	FuncLag       WindowFuncType = "LAG"
	FuncLead      WindowFuncType = "LEAD"
)

type SortOrder string

const (
	SortAsc  SortOrder = "ASC"
	SortDesc SortOrder = "DESC"
)

type NullsOrder string

const (
	NullsFirst NullsOrder = "FIRST"
	NullsLast  NullsOrder = "LAST"
)

type WindowFunction struct {
	Name          WindowFuncType `json:"name"`
	PartitionBy   string         `json:"partition_by,omitempty"`
	OrderBy       string         `json:"order_by,omitempty"`
	Order         SortOrder      `json:"order,omitempty"`
	NullsOrder    NullsOrder     `json:"nulls_order,omitempty"`
	Offset        int            `json:"offset,omitempty"`
	Field         string         `json:"field,omitempty"`
	Alias         string         `json:"alias,omitempty"`
}

type UploadRequest struct {
	Dataset string   `json:"dataset"`
	Data    []Record `json:"data"`
}

type UploadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type QueryRequest struct {
	Dataset   string            `json:"dataset"`
	Functions []WindowFunction  `json:"functions"`
}

type QueryResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message,omitempty"`
	Data    []Record `json:"data,omitempty"`
}

type DatasetInfo struct {
	Name string `json:"name"`
	Size int    `json:"size"`
}

type ListResponse struct {
	Success bool          `json:"success"`
	Datasets []DatasetInfo `json:"datasets,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func (r Record) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}(r))
}

func (r *Record) UnmarshalJSON(data []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	*r = Record(m)
	return nil
}
