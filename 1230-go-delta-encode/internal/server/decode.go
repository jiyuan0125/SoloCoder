package server

import (
	"encoding/json"
	"net/http"

	"delta-encode/pkg/api"
	"delta-encode/pkg/delta"
)

type rawDecodeRequest struct {
	DataType api.DataType    `json:"data_type"`
	Encoded  json.RawMessage `json:"encoded"`
}

func (s *Server) decode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var raw rawDecodeRequest
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	switch raw.DataType {
	case api.DataTypeInt:
		var enc api.IntEncodeResponse
		if err := json.Unmarshal(raw.Encoded, &enc); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid encoded int data: "+err.Error())
			return
		}
		decoded := delta.DecodeInt(delta.IntEncoded{
			Base:   enc.Base,
			Deltas: enc.Deltas,
		})
		writeJSON(w, http.StatusOK, decoded)

	case api.DataTypeFloat:
		var enc api.FloatEncodeResponse
		if err := json.Unmarshal(raw.Encoded, &enc); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid encoded float data: "+err.Error())
			return
		}
		decoded := delta.DecodeFloat(delta.FloatEncoded{
			Base:   enc.Base,
			Deltas: enc.Deltas,
		})
		writeJSON(w, http.StatusOK, decoded)

	default:
		writeError(w, http.StatusBadRequest, "Invalid data_type: must be 'int' or 'float'")
	}
}
