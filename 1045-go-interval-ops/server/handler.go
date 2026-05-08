package main

import (
	"encoding/json"
	"intervalops/common"
	"intervalops/interval"
	"net/http"
	"time"
)

type Server struct {
	store *BatchStore
}

func NewServer() *Server {
	return &Server{
		store: NewBatchStore(),
	}
}

func (s *Server) handleOperation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.OperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, req.BatchID, "invalid request body: "+err.Error())
		return
	}

	result, err := executeOperation(req)
	if err != nil {
		writeErrorResponse(w, req.BatchID, err.Error())
		return
	}

	if req.BatchID != "" {
		s.store.Save(req.BatchID, req, result)
	}

	writeSuccessResponse(w, req.BatchID, result)
}

func (s *Server) handleQueryBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var q common.QueryBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&q); err != nil {
		writeJSON(w, common.QueryBatchResponse{Found: false, Error: "invalid request body: " + err.Error()})
		return
	}

	data, ok := s.store.Get(q.BatchID)
	if !ok {
		writeJSON(w, common.QueryBatchResponse{Found: false})
		return
	}

	writeJSON(w, common.QueryBatchResponse{Found: true, Data: data})
}

func executeOperation(req common.OperationRequest) (common.OperationResult, error) {
	switch req.IntervalType {
	case common.TypeInteger:
		return executeIntegerOperation(req)
	case common.TypeTime:
		return executeTimeOperation(req)
	default:
		return common.OperationResult{}, errInvalidIntervalType
	}
}

func executeIntegerOperation(req common.OperationRequest) (common.OperationResult, error) {
	a := toCoreIntegerIntervals(req.IntegersA)
	b := toCoreIntegerIntervals(req.IntegersB)

	var result []*interval.IntegerInterval
	var err error

	switch req.Operation {
	case common.OpMerge:
		result = interval.MergeIntegers(a)
	case common.OpIntersect:
		result = interval.IntersectionIntegers(a, b)
	case common.OpDifference:
		result = interval.DifferenceIntegers(a, b)
	case common.OpQuery:
		if req.IntegerPoint == nil {
			return common.OperationResult{}, errMissingPoint
		}
		result = interval.QueryPointIntegers(a, *req.IntegerPoint)
	default:
		return common.OperationResult{}, errInvalidOperation
	}

	return common.OperationResult{
		Integers: toCommonIntegerIntervals(result),
	}, err
}

func executeTimeOperation(req common.OperationRequest) (common.OperationResult, error) {
	loc := time.UTC
	if req.Timezone != "" {
		var err error
		loc, err = time.LoadLocation(req.Timezone)
		if err != nil {
			return common.OperationResult{}, errInvalidTimezone
		}
	}

	a := toCoreTimeIntervals(req.TimesA, loc)
	b := toCoreTimeIntervals(req.TimesB, loc)

	var result []*interval.TimeInterval
	var err error

	switch req.Operation {
	case common.OpMerge:
		result = interval.MergeTimes(a)
	case common.OpIntersect:
		result = interval.IntersectionTimes(a, b)
	case common.OpDifference:
		result = interval.DifferenceTimes(a, b)
	case common.OpQuery:
		if req.TimePoint == nil {
			return common.OperationResult{}, errMissingPoint
		}
		result = interval.QueryPointTimes(a, *req.TimePoint)
	default:
		return common.OperationResult{}, errInvalidOperation
	}

	return common.OperationResult{
		Times: toCommonTimeIntervals(result),
	}, err
}

func toCoreIntegerIntervals(ci []common.IntegerInterval) []*interval.IntegerInterval {
	res := make([]*interval.IntegerInterval, len(ci))
	for i, iv := range ci {
		res[i] = interval.NewIntegerInterval(iv.Min, iv.Max)
	}
	return res
}

func toCommonIntegerIntervals(ci []*interval.IntegerInterval) []common.IntegerInterval {
	res := make([]common.IntegerInterval, len(ci))
	for i, iv := range ci {
		res[i] = common.IntegerInterval{Min: iv.Min, Max: iv.Max}
	}
	return res
}

func toCoreTimeIntervals(ct []common.TimeInterval, loc *time.Location) []*interval.TimeInterval {
	res := make([]*interval.TimeInterval, len(ct))
	for i, iv := range ct {
		res[i] = interval.NewTimeInterval(iv.Start, iv.End, loc)
	}
	return res
}

func toCommonTimeIntervals(ct []*interval.TimeInterval) []common.TimeInterval {
	res := make([]common.TimeInterval, len(ct))
	for i, iv := range ct {
		res[i] = common.TimeInterval{Start: iv.Start, End: iv.End}
	}
	return res
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func writeSuccessResponse(w http.ResponseWriter, batchID string, result common.OperationResult) {
	writeJSON(w, common.OperationResponse{
		BatchID: batchID,
		Success: true,
		Result:  result,
	})
}

func writeErrorResponse(w http.ResponseWriter, batchID string, errMsg string) {
	writeJSON(w, common.OperationResponse{
		BatchID: batchID,
		Success: false,
		Error:   errMsg,
	})
}
