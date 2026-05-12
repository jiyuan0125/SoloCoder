package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"protocol-converter/config"
	"protocol-converter/converter"
)

func ConvertHandler(w http.ResponseWriter, r *http.Request) {
	rule := config.GetPathRule(r.URL.Path, r.Method)
	if rule == nil {
		http.Error(w, "No conversion rule found", http.StatusNotFound)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var input map[string]interface{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &input); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
	} else {
		input = make(map[string]interface{})
	}

	convertedRequest := converter.ConvertRequest(input, rule.Request, r.URL.Path)

	internalURL := os.Getenv("INTERNAL_SERVICE_URL")
	if internalURL == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(convertedRequest)
		return
	}

	internalResp, err := forwardToInternal(internalURL+r.URL.Path, r.Method, convertedRequest, r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal service error: %v", err), http.StatusBadGateway)
		return
	}
	defer internalResp.Body.Close()

	respBody, err := io.ReadAll(internalResp.Body)
	if err != nil {
		http.Error(w, "Failed to read internal response", http.StatusBadGateway)
		return
	}

	var internalData map[string]interface{}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &internalData); err != nil {
			http.Error(w, "Invalid JSON from internal service", http.StatusBadGateway)
			return
		}
	} else {
		internalData = make(map[string]interface{})
	}

	convertedResponse := converter.ConvertResponse(internalData, rule.Response, r.URL.Path)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(internalResp.StatusCode)
	json.NewEncoder(w).Encode(convertedResponse)
}

func forwardToInternal(url, method string, data map[string]interface{}, originalReq *http.Request) (*http.Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	for key, values := range originalReq.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	return client.Do(req)
}
