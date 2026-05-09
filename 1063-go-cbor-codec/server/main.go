package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/example/cbor-codec/api"
	"github.com/example/cbor-codec/cbor"
)

const port = ":8080"

func main() {
	http.HandleFunc("/encode", encodeHandler)
	http.HandleFunc("/decode", decodeHandler)
	http.HandleFunc("/json-to-cbor", jsonToCBORHandler)
	http.HandleFunc("/cbor-to-json", cborToJSONHandler)

	fmt.Printf("CBOR server starting on %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}

func encodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	var req api.EncodeRequest
	if err := json.Unmarshal(body, &req); err != nil {
		sendError(w, err.Error())
		return
	}

	value, err := cbor.FromJSON([]byte(req.InputJSON))
	if err != nil {
		sendError(w, err.Error())
		return
	}

	cborData, err := cbor.Marshal(value)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	resp := api.EncodeResponse{
		Success:    true,
		OutputCBOR: hex.EncodeToString(cborData),
	}
	sendResponse(w, resp)
}

func decodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	var req api.DecodeRequest
	if err := json.Unmarshal(body, &req); err != nil {
		sendError(w, err.Error())
		return
	}

	cborData, err := hex.DecodeString(req.InputCBOR)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	value, err := cbor.Unmarshal(cborData)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	jsonData, err := cbor.ToJSON(value)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	resp := api.DecodeResponse{
		Success:    true,
		OutputJSON: string(jsonData),
	}
	sendResponse(w, resp)
}

func jsonToCBORHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	var req api.ConvertJSONToCBORRequest
	if err := json.Unmarshal(body, &req); err != nil {
		sendError(w, err.Error())
		return
	}

	value, err := cbor.FromJSON([]byte(req.InputJSON))
	if err != nil {
		sendError(w, err.Error())
		return
	}

	cborData, err := cbor.Marshal(value)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	resp := api.ConvertJSONToCBORResponse{
		Success:    true,
		OutputCBOR: hex.EncodeToString(cborData),
	}
	sendResponse(w, resp)
}

func cborToJSONHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	var req api.ConvertCBORToJSONRequest
	if err := json.Unmarshal(body, &req); err != nil {
		sendError(w, err.Error())
		return
	}

	cborData, err := hex.DecodeString(req.InputCBOR)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	value, err := cbor.Unmarshal(cborData)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	jsonData, err := cbor.ToJSON(value)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	resp := api.ConvertCBORToJSONResponse{
		Success:    true,
		OutputJSON: string(jsonData),
	}
	sendResponse(w, resp)
}

func sendResponse(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func sendError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	resp := map[string]interface{}{
		"success":       false,
		"error_message": msg,
	}
	json.NewEncoder(w).Encode(resp)
}
