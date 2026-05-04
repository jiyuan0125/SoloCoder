package common

import (
	"querybuilder/pkg/querybuilder"
)

type ConditionRequest struct {
	Type      querybuilder.ConditionType `json:"type"`
	Field     string                      `json:"field"`
	Value     interface{}                 `json:"value,omitempty"`
	Start     interface{}                 `json:"start,omitempty"`
	End       interface{}                 `json:"end,omitempty"`
	Values    []interface{}               `json:"values,omitempty"`
	LikeType  querybuilder.LikeType       `json:"like_type,omitempty"`
}

type SortRequest struct {
	Field string                `json:"field"`
	Order querybuilder.SortOrder `json:"order"`
}

type QueryRequest struct {
	Conditions []ConditionRequest `json:"conditions"`
	SortOrders []SortRequest      `json:"sort_orders"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
}

type QueryResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message,omitempty"`
	Query   querybuilder.Query `json:"query,omitempty"`
}

func NewSuccessResponse(query querybuilder.Query) QueryResponse {
	return QueryResponse{
		Success: true,
		Query:   query,
	}
}

func NewErrorResponse(message string) QueryResponse {
	return QueryResponse{
		Success: false,
		Message: message,
	}
}
