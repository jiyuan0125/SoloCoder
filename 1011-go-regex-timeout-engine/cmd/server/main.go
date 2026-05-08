package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"regex-timeout-engine/internal/protocol"
	"regex-timeout-engine/pkg/regexengine"
)

func main() {
	engine := regexengine.New()

	http.HandleFunc("/regex/compile", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req protocol.CompileRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, protocol.CompileResponse{
				Success: false,
				Error:   "invalid request body",
			})
			return
		}

		timeout := regexengine.DefaultTimeout
		if req.CompileTimeMs > 0 {
			timeout = time.Duration(req.CompileTimeMs) * time.Millisecond
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		regexID, err := engine.Compile(ctx, req.Pattern)
		if err != nil {
			resp := protocol.CompileResponse{
				Success: false,
				Error:   err.Error(),
			}
			if err == regexengine.ErrTimeout {
				writeJSON(w, http.StatusRequestTimeout, resp)
			} else {
				writeJSON(w, http.StatusBadRequest, resp)
			}
			return
		}

		writeJSON(w, http.StatusOK, protocol.CompileResponse{
			RegexID: regexID,
			Success: true,
		})
	})

	http.HandleFunc("/regex/match", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req protocol.MatchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, protocol.MatchResponse{
				Success: false,
				Error:   "invalid request body",
			})
			return
		}

		timeout := regexengine.DefaultTimeout
		if req.MatchTimeMs > 0 {
			timeout = time.Duration(req.MatchTimeMs) * time.Millisecond
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		indices, err := engine.Match(ctx, req.RegexID, req.Text)
		if err != nil {
			resp := protocol.MatchResponse{
				Success: false,
				Error:   err.Error(),
			}
			if err == regexengine.ErrTimeout {
				resp.Timeout = true
				writeJSON(w, http.StatusRequestTimeout, resp)
			} else if err == regexengine.ErrRegexNotFound {
				writeJSON(w, http.StatusNotFound, resp)
			} else {
				writeJSON(w, http.StatusBadRequest, resp)
			}
			return
		}

		matches := make([]protocol.MatchResult, 0, len(indices))
		for _, idx := range indices {
			if len(idx) < 2 {
				continue
			}
			result := protocol.MatchResult{
				Match: req.Text[idx[0]:idx[1]],
				Start: idx[0],
				End:   idx[1],
			}
			for i := 2; i < len(idx); i += 2 {
				if idx[i] >= 0 && idx[i+1] >= 0 {
					result.Groups = append(result.Groups, req.Text[idx[i]:idx[i+1]])
				} else {
					result.Groups = append(result.Groups, "")
				}
			}
			matches = append(matches, result)
		}

		writeJSON(w, http.StatusOK, protocol.MatchResponse{
			Matches: matches,
			Success: true,
		})
	})

	http.HandleFunc("/regex/cache", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		purged := engine.Purge()
		writeJSON(w, http.StatusOK, protocol.PurgeResponse{
			Success: true,
			Purged:  purged,
		})
	})

	addr := ":8080"
	if envAddr := os.Getenv("ADDR"); envAddr != "" {
		addr = envAddr
	}

	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
