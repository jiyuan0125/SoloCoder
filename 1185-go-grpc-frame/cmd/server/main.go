package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"grpc-parser/internal/parser"
	"grpc-parser/pkg/api"
)

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func convertParseResult(result parser.ParseResult) api.ParseResponse {
	response := api.ParseResponse{
		Success:      len(result.Errors) == 0,
		HTTP2Frames:  make([]api.HTTP2FrameInfo, len(result.HTTP2Frames)),
		TotalStreams: result.TotalStreams,
		Errors:       result.Errors,
	}

	for i, frame := range result.HTTP2Frames {
		frameInfo := api.HTTP2FrameInfo{
			Type:          frame.Type,
			StreamID:      frame.StreamID,
			Flags:         frame.Flags,
			Length:        frame.Length,
			HasEndStream:  frame.HasEndStream,
			HasEndHeaders: frame.HasEndHeaders,
			ParseError:    frame.ParseError,
		}

		if frame.Headers != nil {
			frameInfo.Headers = &api.HeadersInfo{
				Headers:   frame.Headers.Headers,
				IsTrailer: frame.Headers.IsTrailer,
				IsGRPC:    frame.Headers.IsGRPC,
			}
		}

		if len(frame.GRPCFrames) > 0 {
			frameInfo.GRPCFrames = make([]api.GRPCFrameInfo, len(frame.GRPCFrames))
			for j, gf := range frame.GRPCFrames {
				grpcInfo := api.GRPCFrameInfo{
					Compressed: gf.Compressed,
					Length:     gf.Length,
					RawMessage: hex.EncodeToString(gf.RawMessage),
					ParseError: gf.ParseError,
				}

				if len(gf.Protobuf) > 0 {
					grpcInfo.Protobuf = make([]api.ProtobufFieldInfo, len(gf.Protobuf))
					for k, pf := range gf.Protobuf {
						grpcInfo.Protobuf[k] = api.ProtobufFieldInfo{
							FieldNumber: pf.FieldNumber,
							WireType:    pf.WireType,
							Value:       pf.Value,
						}
					}
				}

				frameInfo.GRPCFrames[j] = grpcInfo
			}
		}

		response.HTTP2Frames[i] = frameInfo
	}

	return response
}

func handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.ErrorResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
			Success: false,
			Error:   "failed to read request body",
		})
		return
	}
	defer r.Body.Close()

	if len(body) == 0 {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
			Success: false,
			Error:   "empty request body",
		})
		return
	}

	var req api.ParseRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
			Success: false,
			Error:   "invalid JSON: " + err.Error(),
		})
		return
	}

	var data []byte

	if len(req.Bytes) > 0 {
		data = req.Bytes
	} else if req.Hex != "" {
		hexStr := strings.ReplaceAll(strings.ReplaceAll(req.Hex, " ", ""), "\n", "")
		data, err = hex.DecodeString(hexStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
				Success: false,
				Error:   "invalid hex data: " + err.Error(),
			})
			return
		}
	} else {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
			Success: false,
			Error:   "no data provided (bytes or hex field required)",
		})
		return
	}

	result := parser.Parse(data)
	response := convertParseResult(result)

	writeJSON(w, http.StatusOK, response)
}

func handleEncode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.ErrorResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
			Success: false,
			Error:   "failed to read request body",
		})
		return
	}
	defer r.Body.Close()

	if len(body) == 0 {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
			Success: false,
			Error:   "empty request body",
		})
		return
	}

	var req api.EncodeRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
			Success: false,
			Error:   "invalid JSON: " + err.Error(),
		})
		return
	}

	frames := make([]parser.HTTP2Frame, len(req.HTTP2Frames))
	for i, f := range req.HTTP2Frames {
		payload := f.Payload
		if len(payload) == 0 && f.PayloadHex != "" {
			hexStr := strings.ReplaceAll(strings.ReplaceAll(f.PayloadHex, " ", ""), "\n", "")
			decoded, decodeErr := hex.DecodeString(hexStr)
			if decodeErr != nil {
				writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
					Success: false,
					Error:   "invalid payloadHex: " + decodeErr.Error(),
				})
				return
			}
			payload = decoded
		}

		frames[i] = parser.HTTP2Frame{
			Length:   uint32(len(payload)),
			Type:     f.Type,
			Flags:    f.Flags,
			StreamID: f.StreamID,
			Payload:  payload,
		}
	}

	result := parser.Encode(frames)

	response := api.EncodeResponse{
		Success: len(result.Errors) == 0,
		Bytes:   result.Bytes,
		Hex:     hex.EncodeToString(result.Bytes),
		Errors:  result.Errors,
	}

	writeJSON(w, http.StatusOK, response)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func main() {
	var port string

	flag.StringVar(&port, "port", "", "HTTP server port (default: 8080)")
	flag.Parse()

	if port == "" {
		port = getEnv("GRPC_PARSER_PORT", "8204")
	}

	http.HandleFunc("/parse", handleParse)
	http.HandleFunc("/encode", handleEncode)
	http.HandleFunc("/health", handleHealth)

	fmt.Printf("gRPC Parser Server listening on port %s...\n", port)
	fmt.Printf("Endpoints:\n")
	fmt.Printf("  POST /parse  - Parse gRPC over HTTP/2 data\n")
	fmt.Printf("  POST /encode - Encode HTTP/2 frames\n")
	fmt.Printf("  GET  /health - Health check\n")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
