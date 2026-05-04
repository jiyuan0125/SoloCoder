package main

import (
	"encoding/json"
	"net/http"

	"semver-server/api"
	"semver-server/semver"
)

func main() {
	http.HandleFunc("/parse", parseHandler)
	http.HandleFunc("/compare", compareHandler)
	http.HandleFunc("/range", rangeHandler)
	http.HandleFunc("/increment", incrementHandler)

	http.ListenAndServe(":8080", nil)
}

func parseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(api.ParseResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req api.ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(api.ParseResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	version, err := semver.Parse(req.Version)
	if err != nil {
		json.NewEncoder(w).Encode(api.ParseResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(api.ParseResponse{
		Success: true,
		Version: &api.Version{
			Major:      version.Major,
			Minor:      version.Minor,
			Patch:      version.Patch,
			PreRelease: version.PreRelease,
			String:     version.String(),
		},
	})
}

func compareHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(api.CompareResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req api.CompareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(api.CompareResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	v1, err := semver.Parse(req.Version1)
	if err != nil {
		json.NewEncoder(w).Encode(api.CompareResponse{
			Success: false,
			Error:   "invalid version1: " + err.Error(),
		})
		return
	}

	v2, err := semver.Parse(req.Version2)
	if err != nil {
		json.NewEncoder(w).Encode(api.CompareResponse{
			Success: false,
			Error:   "invalid version2: " + err.Error(),
		})
		return
	}

	result := semver.Compare(v1, v2)
	json.NewEncoder(w).Encode(api.CompareResponse{
		Success: true,
		Result:  result,
	})
}

func rangeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(api.RangeCheckResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req api.RangeCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(api.RangeCheckResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	version, err := semver.Parse(req.Version)
	if err != nil {
		json.NewEncoder(w).Encode(api.RangeCheckResponse{
			Success: false,
			Error:   "invalid version: " + err.Error(),
		})
		return
	}

	versionRange, err := semver.ParseRange(req.Range)
	if err != nil {
		json.NewEncoder(w).Encode(api.RangeCheckResponse{
			Success: false,
			Error:   "invalid range: " + err.Error(),
		})
		return
	}

	inRange := version.IsInRange(versionRange)
	json.NewEncoder(w).Encode(api.RangeCheckResponse{
		Success: true,
		InRange: inRange,
	})
}

func incrementHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(api.IncrementResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req api.IncrementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(api.IncrementResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	version, err := semver.Parse(req.Version)
	if err != nil {
		json.NewEncoder(w).Encode(api.IncrementResponse{
			Success: false,
			Error:   "invalid version: " + err.Error(),
		})
		return
	}

	switch req.Part {
	case "major":
		version.IncrementMajor()
	case "minor":
		version.IncrementMinor()
	case "patch":
		version.IncrementPatch()
	default:
		json.NewEncoder(w).Encode(api.IncrementResponse{
			Success: false,
			Error:   "invalid part: must be major, minor, or patch",
		})
		return
	}

	json.NewEncoder(w).Encode(api.IncrementResponse{
		Success: true,
		Version: &api.Version{
			Major:      version.Major,
			Minor:      version.Minor,
			Patch:      version.Patch,
			PreRelease: version.PreRelease,
			String:     version.String(),
		},
	})
}
