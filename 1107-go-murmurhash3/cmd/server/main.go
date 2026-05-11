package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"murmurhash3/pkg/api"
	"murmurhash3/pkg/murmurhash3"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, api.ErrorResponse{Error: msg})
}

func hash32Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.HashRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	hash := murmurhash3.Hash32String(req.Data, req.Seed)
	hexStr := fmt.Sprintf("%08x", hash)

	writeJSON(w, http.StatusOK, api.Hash32Response{
		Hash: hash,
		Hex:  hexStr,
		Seed: req.Seed,
	})
}

func hash128Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.HashRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	h1, h2 := murmurhash3.Hash128String(req.Data, req.Seed)
	highBytes := make([]byte, 8)
	lowBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		highBytes[7-i] = byte(h1 >> (i * 8))
		lowBytes[7-i] = byte(h2 >> (i * 8))
	}
	hexStr := hex.EncodeToString(highBytes) + hex.EncodeToString(lowBytes)

	writeJSON(w, http.StatusOK, api.Hash128Response{
		High: h1,
		Low:  h2,
		Hex:  hexStr,
		Seed: req.Seed,
	})
}

func distributionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.DistributionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Data) == 0 {
		writeError(w, http.StatusBadRequest, "no data provided")
		return
	}

	data := make([][]byte, len(req.Data))
	for i, s := range req.Data {
		data[i] = []byte(s)
	}

	bucketCount := req.BucketCount
	if bucketCount <= 0 {
		bucketCount = 256
	}

	var report murmurhash3.DistributionReport
	hashType := strings.ToLower(req.HashType)
	if hashType == "128" || hashType == "x86_128" {
		report = murmurhash3.AnalyzeDistribution128(data, req.Seed, bucketCount)
	} else {
		report = murmurhash3.AnalyzeDistribution32(data, req.Seed, bucketCount)
	}

	writeJSON(w, http.StatusOK, api.DistributionReport{
		Min:          report.Min,
		Max:          report.Max,
		Mean:         report.Mean,
		StdDev:       report.StdDev,
		Variance:     report.Variance,
		ChiSquare:    report.ChiSquare,
		UniformScore: report.UniformScore,
		BucketCount:  report.BucketCount,
		SampleCount:  report.SampleCount,
	})
}

func main() {
	http.HandleFunc("/hash/32", hash32Handler)
	http.HandleFunc("/hash/128", hash128Handler)
	http.HandleFunc("/distribution", distributionHandler)

	port := 8410
	fmt.Printf("Server listening on :%d\n", port)
	if err := http.ListenAndServe(":"+strconv.Itoa(port), nil); err != nil {
		panic(err)
	}
}
