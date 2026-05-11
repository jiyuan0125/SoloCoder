package server

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"kmeans-cluster/common"
	"kmeans-cluster/kmeans"
)

const DefaultMaxK = 10

func toKmeansPoints(points []common.Point) []kmeans.Point {
	result := make([]kmeans.Point, len(points))
	for i, p := range points {
		result[i] = kmeans.Point(p)
	}
	return result
}

func toCommonPoints(points []kmeans.Point) []common.Point {
	result := make([]common.Point, len(points))
	for i, p := range points {
		result[i] = common.Point(p)
	}
	return result
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(common.ErrorResponse{
		Success: false,
		Error:   message,
	})
}

func parsePointsFromRequest(r *http.Request) ([]kmeans.Point, error) {
	contentType := r.Header.Get("Content-Type")

	if strings.Contains(contentType, "multipart/form-data") {
		file, _, err := r.FormFile("file")
		if err != nil {
			return nil, fmt.Errorf("failed to get file: %v", err)
		}
		defer file.Close()

		filename := r.FormValue("filename")
		if strings.HasSuffix(strings.ToLower(filename), ".csv") {
			return parseCSV(file)
		}
		return parseJSONPoints(file)
	}

	if strings.Contains(contentType, "text/csv") || strings.HasSuffix(r.URL.Path, ".csv") {
		return parseCSV(r.Body)
	}

	return parseJSONRequest(r.Body)
}

func parseCSV(r io.Reader) ([]kmeans.Point, error) {
	reader := csv.NewReader(r)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %v", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no data in CSV")
	}

	points := make([]kmeans.Point, 0, len(records))
	for i, record := range records {
		point := make(kmeans.Point, 0, len(record))
		for j, val := range record {
			val = strings.TrimSpace(val)
			if val == "" {
				return nil, fmt.Errorf("empty value at row %d, column %d", i+1, j+1)
			}
			f, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid number '%s' at row %d, column %d: %v", val, i+1, j+1, err)
			}
			point = append(point, f)
		}
		if len(point) == 0 {
			return nil, fmt.Errorf("empty row at row %d", i+1)
		}
		if i > 0 && len(point) != len(points[0]) {
			return nil, fmt.Errorf("dimension mismatch at row %d: expected %d, got %d", i+1, len(points[0]), len(point))
		}
		points = append(points, point)
	}

	return points, nil
}

func parseJSONPoints(r io.Reader) ([]kmeans.Point, error) {
	var points []kmeans.Point
	if err := json.NewDecoder(r).Decode(&points); err != nil {
		return nil, fmt.Errorf("failed to parse JSON points: %v", err)
	}
	return points, nil
}

func parseJSONRequest(r io.Reader) ([]kmeans.Point, error) {
	body, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %v", err)
	}

	var req struct {
		Points []kmeans.Point `json:"points"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	if len(req.Points) == 0 {
		var points []kmeans.Point
		if err := json.Unmarshal(body, &points); err == nil && len(points) > 0 {
			return points, nil
		}
		return nil, fmt.Errorf("no points provided")
	}

	return req.Points, nil
}

func parseClusterRequest(r *http.Request) (*common.ClusterRequest, []kmeans.Point, error) {
	var req common.ClusterRequest

	contentType := r.Header.Get("Content-Type")

	if strings.Contains(contentType, "multipart/form-data") {
		points, err := parsePointsFromRequest(r)
		if err != nil {
			return nil, nil, err
		}

		if kStr := r.FormValue("k"); kStr != "" {
			k, err := strconv.Atoi(kStr)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid k value: %v", err)
			}
			req.K = &k
		}

		if autoKStr := r.FormValue("auto_k"); autoKStr != "" {
			req.AutoK, _ = strconv.ParseBool(autoKStr)
		}

		if maxKStr := r.FormValue("max_k"); maxKStr != "" {
			maxK, err := strconv.Atoi(maxKStr)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid max_k value: %v", err)
			}
			req.MaxK = &maxK
		}

		if iterStr := r.FormValue("max_iterations"); iterStr != "" {
			iter, err := strconv.Atoi(iterStr)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid max_iterations value: %v", err)
			}
			req.MaxIterations = &iter
		}

		if epsStr := r.FormValue("epsilon"); epsStr != "" {
			eps, err := strconv.ParseFloat(epsStr, 64)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid epsilon value: %v", err)
			}
			req.Epsilon = &eps
		}

		req.Points = toCommonPoints(points)
		return &req, points, nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read request body: %v", err)
	}

	if err := json.Unmarshal(body, &req); err != nil {
		return nil, nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	if len(req.Points) == 0 {
		return nil, nil, fmt.Errorf("no points provided")
	}

	return &req, toKmeansPoints(req.Points), nil
}

func parseWCSSRequest(r *http.Request) (*common.WCSSRequest, []kmeans.Point, error) {
	var req common.WCSSRequest

	contentType := r.Header.Get("Content-Type")

	if strings.Contains(contentType, "multipart/form-data") {
		points, err := parsePointsFromRequest(r)
		if err != nil {
			return nil, nil, err
		}

		maxKStr := r.FormValue("max_k")
		if maxKStr == "" {
			return nil, nil, fmt.Errorf("max_k is required")
		}
		maxK, err := strconv.Atoi(maxKStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid max_k value: %v", err)
		}
		req.MaxK = maxK

		if iterStr := r.FormValue("max_iterations"); iterStr != "" {
			iter, err := strconv.Atoi(iterStr)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid max_iterations value: %v", err)
			}
			req.MaxIterations = &iter
		}

		if epsStr := r.FormValue("epsilon"); epsStr != "" {
			eps, err := strconv.ParseFloat(epsStr, 64)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid epsilon value: %v", err)
			}
			req.Epsilon = &eps
		}

		req.Points = toCommonPoints(points)
		return &req, points, nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read request body: %v", err)
	}

	if err := json.Unmarshal(body, &req); err != nil {
		return nil, nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	if len(req.Points) == 0 {
		return nil, nil, fmt.Errorf("no points provided")
	}

	return &req, toKmeansPoints(req.Points), nil
}

func ClusterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	req, points, err := parseClusterRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if len(points) == 0 {
		writeError(w, http.StatusBadRequest, "No data points provided")
		return
	}

	maxIterations := kmeans.DefaultMaxIterations
	if req.MaxIterations != nil && *req.MaxIterations > 0 {
		maxIterations = *req.MaxIterations
	}

	epsilon := kmeans.DefaultEpsilon
	if req.Epsilon != nil && *req.Epsilon > 0 {
		epsilon = *req.Epsilon
	}

	var k int
	var wcssMap map[int]float64

	if req.AutoK {
		maxK := DefaultMaxK
		if req.MaxK != nil && *req.MaxK > 0 {
			maxK = *req.MaxK
		}
		if maxK > len(points) {
			maxK = len(points)
		}

		var err error
		wcssMap, err = kmeans.ElbowMethod(points, maxK, maxIterations, epsilon)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to compute WCSS: %v", err))
			return
		}

		k = kmeans.FindOptimalK(wcssMap)
	} else if req.K != nil {
		k = *req.K
		if k <= 0 {
			writeError(w, http.StatusBadRequest, "K must be positive")
			return
		}
		if k > len(points) {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("K cannot be larger than number of points (%d)", len(points)))
			return
		}
	} else {
		writeError(w, http.StatusBadRequest, "Either k or auto_k must be specified")
		return
	}

	result, err := kmeans.KMeans(points, k, maxIterations, epsilon)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("K-Means failed: %v", err))
		return
	}

	clusterStats, err := kmeans.ClusterStats(points, result.Labels, result.Centers)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to compute cluster stats: %v", err))
		return
	}

	commonClusterInfos := make([]common.ClusterInfo, len(clusterStats))
	for i, cs := range clusterStats {
		commonClusterInfos[i] = common.ClusterInfo{
			Index:       cs.Index,
			Center:      common.Point(cs.Center),
			PointCount:  cs.PointCount,
			AvgDistance: cs.AvgDistance,
		}
	}

	response := common.ClusterResponse{
		Success:    true,
		Labels:     result.Labels,
		Centers:    make([]common.Point, len(result.Centers)),
		K:          k,
		Iterations: result.Iterations,
		Converged:  result.Converged,
		Clusters:   commonClusterInfos,
	}

	for i, c := range result.Centers {
		response.Centers[i] = common.Point(c)
	}

	if req.AutoK {
		response.Message = fmt.Sprintf("Auto-selected K=%d using Elbow Method", k)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func WCSSHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	req, points, err := parseWCSSRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if len(points) == 0 {
		writeError(w, http.StatusBadRequest, "No data points provided")
		return
	}

	if req.MaxK <= 0 {
		writeError(w, http.StatusBadRequest, "max_k must be positive")
		return
	}

	maxIterations := kmeans.DefaultMaxIterations
	if req.MaxIterations != nil && *req.MaxIterations > 0 {
		maxIterations = *req.MaxIterations
	}

	epsilon := kmeans.DefaultEpsilon
	if req.Epsilon != nil && *req.Epsilon > 0 {
		epsilon = *req.Epsilon
	}

	wcssMap, err := kmeans.ElbowMethod(points, req.MaxK, maxIterations, epsilon)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to compute WCSS: %v", err))
		return
	}

	response := common.WCSSResponse{
		Success: true,
		WCSS:    wcssMap,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
