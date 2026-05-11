package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"crc-checksum/pkg/api"
	"crc-checksum/pkg/crc"
)

func parseHexToUint32(s string) (uint32, error) {
	s = strings.TrimPrefix(strings.TrimPrefix(s, "0x"), "0X")
	if len(s) > 8 {
		return 0, fmt.Errorf("hex value too large for 32-bit: %s", s)
	}
	value, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return 0, err
	}
	return uint32(value), nil
}

func parseHexToUint64(s string) (uint64, error) {
	s = strings.TrimPrefix(strings.TrimPrefix(s, "0x"), "0X")
	if len(s) > 16 {
		return 0, fmt.Errorf("hex value too large for 64-bit: %s", s)
	}
	value, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func decodeData(data string, isBase64 bool) ([]byte, error) {
	if data == "" {
		return []byte{}, nil
	}
	if isBase64 {
		return base64.StdEncoding.DecodeString(data)
	}
	return []byte(data), nil
}

func getCRC32Variant(variant string, custom *api.CustomCRC32Config) (*crc.CRC32Variant, error) {
	if custom != nil {
		poly, err := parseHexToUint32(custom.Polynomial)
		if err != nil {
			return nil, fmt.Errorf("invalid polynomial: %v", err)
		}
		init, err := parseHexToUint32(custom.Init)
		if err != nil {
			return nil, fmt.Errorf("invalid init value: %v", err)
		}
		xorOut, err := parseHexToUint32(custom.XorOut)
		if err != nil {
			return nil, fmt.Errorf("invalid xor_out value: %v", err)
		}

		return &crc.CRC32Variant{
			Name:       "custom",
			Polynomial: poly,
			Init:       init,
			RefIn:      custom.RefIn,
			RefOut:     custom.RefOut,
			XorOut:     xorOut,
		}, nil
	}

	if variant == "" {
		variant = "iso-hdlc"
	}

	v, ok := crc.CRC32Variants[strings.ToLower(variant)]
	if !ok {
		return nil, fmt.Errorf("unknown CRC32 variant: %s", variant)
	}

	return &v, nil
}

func getCRC64Variant(variant string, custom *api.CustomCRC64Config) (*crc.CRC64Variant, error) {
	if custom != nil {
		poly, err := parseHexToUint64(custom.Polynomial)
		if err != nil {
			return nil, fmt.Errorf("invalid polynomial: %v", err)
		}
		init, err := parseHexToUint64(custom.Init)
		if err != nil {
			return nil, fmt.Errorf("invalid init value: %v", err)
		}
		xorOut, err := parseHexToUint64(custom.XorOut)
		if err != nil {
			return nil, fmt.Errorf("invalid xor_out value: %v", err)
		}

		return &crc.CRC64Variant{
			Name:       "custom",
			Polynomial: poly,
			Init:       init,
			RefIn:      custom.RefIn,
			RefOut:     custom.RefOut,
			XorOut:     xorOut,
		}, nil
	}

	if variant == "" {
		variant = "ecma"
	}

	v, ok := crc.CRC64Variants[strings.ToLower(variant)]
	if !ok {
		return nil, fmt.Errorf("unknown CRC64 variant: %s", variant)
	}

	return &v, nil
}

func variantsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	crc32Variants := make(map[string]string)
	for name, v := range crc.CRC32Variants {
		crc32Variants[name] = v.Name
	}

	crc64Variants := make(map[string]string)
	for name, v := range crc.CRC64Variants {
		crc64Variants[name] = v.Name
	}

	response := api.VariantsResponse{
		Success: true,
		CRC32:   crc32Variants,
		CRC64:   crc64Variants,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func calculateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.CalculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(api.CalculateResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	data, err := decodeData(req.Data, req.IsBase64)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(api.CalculateResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to decode data: %v", err),
		})
		return
	}

	var crcResult string

	switch req.Type {
	case api.CRCType32:
		variant, err := getCRC32Variant(req.Variant, req.CustomCRC32)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(api.CalculateResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}

		crcValue, err := crc.CalculateCRC32(data, *variant, nil)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(api.CalculateResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}

		crcResult = crc.FormatCRC32(crcValue)

	case api.CRCType64:
		variant, err := getCRC64Variant(req.Variant, req.CustomCRC64)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(api.CalculateResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}

		crcValue, err := crc.CalculateCRC64(data, *variant, nil)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(api.CalculateResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}

		crcResult = crc.FormatCRC64(crcValue)

	default:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(api.CalculateResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid CRC type: %s, must be 'crc32' or 'crc64'", req.Type),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.CalculateResponse{
		Success: true,
		CRC:     crcResult,
	})
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(api.VerifyResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	data, err := decodeData(req.Data, req.IsBase64)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(api.VerifyResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to decode data: %v", err),
		})
		return
	}

	var match bool
	var actualCRC string

	switch req.Type {
	case api.CRCType32:
		expected, err := parseHexToUint32(req.ExpectedCRC)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(api.VerifyResponse{
				Success: false,
				Error:   fmt.Sprintf("Invalid expected CRC: %v", err),
			})
			return
		}

		variant, err := getCRC32Variant(req.Variant, req.CustomCRC32)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(api.VerifyResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}

		actual, err := crc.CalculateCRC32(data, *variant, nil)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(api.VerifyResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}

		match = actual == expected
		actualCRC = crc.FormatCRC32(actual)

	case api.CRCType64:
		expected, err := parseHexToUint64(req.ExpectedCRC)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(api.VerifyResponse{
				Success: false,
				Error:   fmt.Sprintf("Invalid expected CRC: %v", err),
			})
			return
		}

		variant, err := getCRC64Variant(req.Variant, req.CustomCRC64)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(api.VerifyResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}

		actual, err := crc.CalculateCRC64(data, *variant, nil)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(api.VerifyResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}

		match = actual == expected
		actualCRC = crc.FormatCRC64(actual)

	default:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(api.VerifyResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid CRC type: %s, must be 'crc32' or 'crc64'", req.Type),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.VerifyResponse{
		Success: true,
		Match:   match,
		Actual:  actualCRC,
	})
}

func main() {
	addr := flag.String("addr", ":8303", "Address to listen on")
	flag.Parse()

	http.HandleFunc("/variants", variantsHandler)
	http.HandleFunc("/calculate", calculateHandler)
	http.HandleFunc("/verify", verifyHandler)

	fmt.Printf("CRC Checksum Server listening on %s\n", *addr)
	fmt.Println("Endpoints:")
	fmt.Println("  GET  /variants       - List available CRC variants")
	fmt.Println("  POST /calculate      - Calculate CRC value")
	fmt.Println("  POST /verify         - Verify CRC value")

	if err := http.ListenAndServe(*addr, nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
