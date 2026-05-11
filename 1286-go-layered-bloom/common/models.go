package common

import "layered-bloom/core"

type InsertRequest struct {
	Element string `json:"element"`
}

type InsertResponse struct {
	Success     bool   `json:"success"`
	Inserted    bool   `json:"inserted"`
	LayerIndex  int    `json:"layer_index"`
	WasExisting bool   `json:"was_existing"`
	Message     string `json:"message,omitempty"`
}

type QueryRequest struct {
	Element string `json:"element"`
}

type QueryResponse struct {
	Exists    bool   `json:"exists"`
	Message   string `json:"message,omitempty"`
}

type DeleteRequest struct {
	Element string `json:"element"`
}

type DeleteResponse struct {
	Success   bool   `json:"success"`
	Removed   bool   `json:"removed"`
	Message   string `json:"message,omitempty"`
}

type StatsResponse struct {
	Success    bool              `json:"success"`
	Stats      *core.FilterStats `json:"stats,omitempty"`
	Message    string            `json:"message,omitempty"`
}

type ResetResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
}

type LayerConfig struct {
	Capacity       int     `json:"capacity"`
	HashFunctions  int     `json:"hash_functions"`
	TargetFPR      float64 `json:"target_fpr"`
}

type ConfigRequest struct {
	Layers            []LayerConfig `json:"layers"`
	AutoExpandOnFull  bool          `json:"auto_expand_on_full"`
	WarningFPRFactor  float64       `json:"warning_fpr_factor"`
	MaxAutoLayers     int           `json:"max_auto_layers"`
}

type ConfigResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
}

type BatchInsertRequest struct {
	Elements []string `json:"elements"`
}

type BatchInsertResult struct {
	Element     string `json:"element"`
	Success     bool   `json:"success"`
	Inserted    bool   `json:"inserted"`
	LayerIndex  int    `json:"layer_index"`
	WasExisting bool   `json:"was_existing"`
}

type BatchInsertResponse struct {
	Success bool                `json:"success"`
	Results []BatchInsertResult `json:"results"`
	Total   int                 `json:"total"`
	Inserted int                `json:"inserted"`
	Existing int                `json:"existing"`
	Failed  int                 `json:"failed"`
}

type BatchQueryRequest struct {
	Elements []string `json:"elements"`
}

type BatchQueryResult struct {
	Element string `json:"element"`
	Exists  bool   `json:"exists"`
}

type BatchQueryResponse struct {
	Success   bool               `json:"success"`
	Results   []BatchQueryResult `json:"results"`
	Total     int                `json:"total"`
	Exists    int                `json:"exists"`
	NotExists int                `json:"not_exists"`
}

type BatchDeleteRequest struct {
	Elements []string `json:"elements"`
}

type BatchDeleteResult struct {
	Element string `json:"element"`
	Success bool   `json:"success"`
	Removed bool   `json:"removed"`
}

type BatchDeleteResponse struct {
	Success bool                `json:"success"`
	Results []BatchDeleteResult `json:"results"`
	Total   int                 `json:"total"`
	Removed int                 `json:"removed"`
	NotFound int                `json:"not_found"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
