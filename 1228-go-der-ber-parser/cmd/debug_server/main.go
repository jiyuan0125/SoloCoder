package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"der-ber-parser/api"
	"der-ber-parser/asn1"
)

func main() {
	http.HandleFunc("/debug-encode-der", handleDebugEncodeDER)
	log.Println("Debug server starting on :8085")
	http.ListenAndServe(":8085", nil)
}

func handleDebugEncodeDER(w http.ResponseWriter, r *http.Request) {
	// Read raw body first
	bodyBytes, err := readAll(r.Body)
	if err != nil {
		http.Error(w, "read error", 500)
		return
	}

	fmt.Printf("=== Raw request body ===\n")
	fmt.Printf("%s\n", string(bodyBytes))

	// Now parse
	var req api.EncodeRequest
	err = json.Unmarshal(bodyBytes, &req)
	if err != nil {
		fmt.Printf("JSON parse error: %v\n", err)
		http.Error(w, err.Error(), 400)
		return
	}

	fmt.Printf("\n=== Parsed req ===\n")
	fmt.Printf("req.Mode = %q\n", req.Mode)
	fmt.Printf("req.Data.Type = %q\n", req.Data.Type)
	fmt.Printf("req.Data.Tag = %q\n", req.Data.Tag)
	fmt.Printf("req.Data.TagNumber = %d\n", req.Data.TagNumber)
	fmt.Printf("req.Data.Constructed = %v\n", req.Data.Constructed)
	fmt.Printf("req.Data.Value = %v\n", req.Data.Value)
	fmt.Printf("req.Data.Children = %v\n", req.Data.Children)

	// Try to encode
	hex, err := asn1.EncodeDER(req.Data)
	fmt.Printf("\n=== Encode result ===\n")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Hex: %s\n", hex)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.EncodeResponse{
		Success: err == nil,
		Hex:     hex,
		Error:   fmtErr(err),
	})
}

func readAll(r io.ReadCloser) ([]byte, error) {
	defer r.Close()
	var buf []byte
	tmp := make([]byte, 1024)
	for {
		n, err := r.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			if err == io.EOF {
				return buf, nil
			}
			return buf, err
		}
	}
}

func fmtErr(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
