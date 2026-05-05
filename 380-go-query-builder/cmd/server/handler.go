package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"querybuilder/pkg/common"
	"querybuilder/pkg/querybuilder"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) BuildQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.QueryRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request body: " + err.Error()))
		return
	}

	qb := querybuilder.New()

	for _, cond := range req.Conditions {
		switch cond.Type {
		case querybuilder.ConditionTypeEqual:
			qb = qb.Equal(cond.Field, cond.Value)
		case querybuilder.ConditionTypeRange:
			qb = qb.Range(cond.Field, cond.Start, cond.End)
		case querybuilder.ConditionTypeIn:
			qb = qb.In(cond.Field, cond.Values...)
		case querybuilder.ConditionTypeLike:
			valueStr := interfaceToString(cond.Value)
			likeType := cond.LikeType
			if likeType == "" {
				likeType = querybuilder.LikeTypeBoth
			}
			qb = qb.Like(cond.Field, valueStr, likeType)
		}
	}

	for _, sort := range req.SortOrders {
		qb = qb.Sort(sort.Field, sort.Order)
	}

	qb = qb.Pagination(req.Page, req.PageSize)

	query := qb.Build()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(query))
}

func interfaceToString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		return fmt.Sprintf("%v", val)
	case bool:
		return fmt.Sprintf("%t", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}
