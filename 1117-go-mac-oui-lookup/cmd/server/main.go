package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"mac-oui-lookup/pkg/api"
	"mac-oui-lookup/pkg/macaddr"
	"mac-oui-lookup/pkg/oui"
)

type Server struct {
	ouiDB *oui.Database
}

func NewServer() *Server {
	return &Server{
		ouiDB: oui.NewDatabase(),
	}
}

func (s *Server) LoadOUIDatabase(path string) error {
	if path == "" {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return s.ouiDB.LoadFromReader(file)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func getMACAddress(r *http.Request) string {
	mac := r.URL.Query().Get("mac")
	if mac == "" {
		var req api.ParseRequest
		json.NewDecoder(r.Body).Decode(&req)
		mac = req.MAC
	}
	return mac
}

func (s *Server) handleParse(w http.ResponseWriter, r *http.Request) {
	macStr := getMACAddress(r)
	if macStr == "" {
		writeJSON(w, http.StatusBadRequest, api.ParseResponse{
			Success: false,
			Error:   "missing MAC address",
		})
		return
	}

	mac, err := macaddr.ParseFormat(macStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.ParseResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	types := getMACTypes(mac)

	var vendor string
	if v, err := s.ouiDB.Lookup(mac.Bytes()); err == nil {
		vendor = v
	}

	writeJSON(w, http.StatusOK, api.ParseResponse{
		Success: true,
		MAC: api.MACInfo{
			Address:          mac.String(),
			OUI:              formatOUI(mac.OUI()),
			Vendor:           vendor,
			Types:            types,
			IsMulticast:      mac.IsMulticast(),
			IsLocallyAdmin:   mac.IsLocallyAdministered(),
			IsGloballyUnique: mac.IsGloballyUnique(),
			IsZero:           mac.IsZero(),
			IsBroadcast:      mac.IsBroadcast(),
		},
	})
}

func (s *Server) handleFormat(w http.ResponseWriter, r *http.Request) {
	macStr := r.URL.Query().Get("mac")
	formatStr := r.URL.Query().Get("format")
	if macStr == "" {
		var req api.FormatRequest
		json.NewDecoder(r.Body).Decode(&req)
		macStr = req.MAC
		formatStr = req.Format
	}

	if macStr == "" {
		writeJSON(w, http.StatusBadRequest, api.FormatResponse{
			Success: false,
			Error:   "missing MAC address",
		})
		return
	}

	mac, err := macaddr.ParseFormat(macStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.FormatResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	format := parseFormat(formatStr)
	result := mac.Format(format)

	writeJSON(w, http.StatusOK, api.FormatResponse{
		Success: true,
		Result:  result,
	})
}

func (s *Server) handleOUI(w http.ResponseWriter, r *http.Request) {
	macStr := getMACAddress(r)
	if macStr == "" {
		writeJSON(w, http.StatusBadRequest, api.OUIResponse{
			Success: false,
			Error:   "missing MAC address",
		})
		return
	}

	mac, err := macaddr.ParseFormat(macStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.OUIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	vendor, err := s.ouiDB.Lookup(mac.Bytes())
	if err != nil {
		writeJSON(w, http.StatusOK, api.OUIResponse{
			Success: true,
			OUI:     formatOUI(mac.OUI()),
		})
		return
	}

	writeJSON(w, http.StatusOK, api.OUIResponse{
		Success: true,
		OUI:     formatOUI(mac.OUI()),
		Vendor:  vendor,
	})
}

func (s *Server) handleTypes(w http.ResponseWriter, r *http.Request) {
	macStr := getMACAddress(r)
	if macStr == "" {
		writeJSON(w, http.StatusBadRequest, api.TypesResponse{
			Success: false,
			Error:   "missing MAC address",
		})
		return
	}

	mac, err := macaddr.ParseFormat(macStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.TypesResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	types := getMACTypes(mac)

	writeJSON(w, http.StatusOK, api.TypesResponse{
		Success: true,
		Types:   types,
	})
}

func getMACTypes(mac macaddr.MAC) []string {
	types := []string{}

	if mac.IsZero() {
		types = append(types, "uninitialized")
	}
	if mac.IsBroadcast() {
		types = append(types, "broadcast")
	}
	if mac.IsMulticast() {
		types = append(types, "multicast")
	}
	if mac.IsLocallyAdministered() {
		types = append(types, "locally-administered")
	}
	if mac.IsGloballyUnique() && !mac.IsZero() && !mac.IsBroadcast() && !mac.IsMulticast() {
		types = append(types, "globally-unique")
	}
	if mac.IsUnicast() && !mac.IsZero() && !mac.IsBroadcast() {
		types = append(types, "unicast")
	}

	if len(types) == 0 {
		types = append(types, "unicast")
	}

	return types
}

func formatOUI(oui [3]byte) string {
	return fmt.Sprintf("%02x:%02x:%02x", oui[0], oui[1], oui[2])
}

func parseFormat(s string) macaddr.Format {
	switch strings.ToLower(s) {
	case "dash", "hyphen":
		return macaddr.FormatDash
	case "dot", "cisco":
		return macaddr.FormatDot
	default:
		return macaddr.FormatColon
	}
}

func main() {
	addr := flag.String("addr", ":8080", "HTTP server address")
	ouiFile := flag.String("oui", "", "OUI database file path")
	flag.Parse()

	server := NewServer()

	if *ouiFile != "" {
		if err := server.LoadOUIDatabase(*ouiFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading OUI database: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Loaded %d OUI entries\n", server.ouiDB.Size())
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/parse", server.handleParse)
	mux.HandleFunc("/format", server.handleFormat)
	mux.HandleFunc("/oui", server.handleOUI)
	mux.HandleFunc("/types", server.handleTypes)

	fmt.Printf("Server listening on %s\n", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
