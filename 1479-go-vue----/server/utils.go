package main

import (
	"encoding/json"
	"net/http"

	"github.com/drivingschool/common"
)

func writeJSON(w http.ResponseWriter, statusCode int, response common.Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
