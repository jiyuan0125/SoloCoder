package main

import (
	"encoding/json"
	"go-ip-parser/pkg/common"
	"go-ip-parser/pkg/ipparser"
	"log"
	"net/http"
	"strings"
)

func handleParse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, `{"success": false, "error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req common.ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.ParseResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	addr, err := ipparser.Parse(req.IP)
	if err != nil {
		resp := common.ParseResponse{
			Success:  false,
			Error:    err.Error(),
			Original: req.IP,
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := buildParseResponse(req.IP, addr)
	json.NewEncoder(w).Encode(resp)
}

func handleFormat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, `{"success": false, "error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req common.FormatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.FormatResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	addr, err := ipparser.Parse(req.IP)
	if err != nil {
		resp := common.FormatResponse{
			Success:  false,
			Error:    err.Error(),
			Original: req.IP,
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := common.FormatResponse{
		Success:   true,
		Original:  req.IP,
		Formatted: addr.Standard(),
		Compact:   addr.Canonical(),
	}

	json.NewEncoder(w).Encode(resp)
}

func handleClassify(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, `{"success": false, "error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req common.ClassifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.ClassifyResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	addr, err := ipparser.Parse(req.IP)
	if err != nil {
		resp := common.ClassifyResponse{
			Success: false,
			Error:   err.Error(),
			IP:      req.IP,
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := buildClassifyResponse(req.IP, addr)
	json.NewEncoder(w).Encode(resp)
}

func buildParseResponse(original string, addr *ipparser.Address) common.ParseResponse {
	resp := common.ParseResponse{
		Success:       true,
		Original:      original,
		Version:       addr.Version(),
		Standard:      addr.Standard(),
		Canonical:     addr.Canonical(),
		IsLoopback:    addr.IsLoopback(),
		IsPrivate:     addr.IsPrivate(),
		IsMulticast:   addr.IsMulticast(),
		IsLinkLocal:   addr.IsLinkLocal(),
		IsUnspecified: addr.IsUnspecified(),
	}

	if addr.IsIPv4() {
		octets := addr.IPv4Octets()
		resp.IPv4Octets = [4]uint8{uint8(octets[0]), uint8(octets[1]), uint8(octets[2]), uint8(octets[3])}
		class := addr.GetIPv4Class()
		resp.IPv4Class = &class
	} else if addr.IsIPv6() {
		groups := addr.IPv6Groups()
		resp.IPv6Groups = groups
		ipv6Type := addr.GetIPv6Type()
		resp.IPv6Type = &ipv6Type
	}

	return resp
}

func buildClassifyResponse(original string, addr *ipparser.Address) common.ClassifyResponse {
	resp := common.ClassifyResponse{
		Success:       true,
		IP:            original,
		Version:       addr.Version(),
		IsLoopback:    addr.IsLoopback(),
		IsPrivate:     addr.IsPrivate(),
		IsMulticast:   addr.IsMulticast(),
		IsLinkLocal:   addr.IsLinkLocal(),
		IsUnspecified: addr.IsUnspecified(),
	}

	var desc []string

	if addr.IsUnspecified() {
		desc = append(desc, "unspecified (wildcard)")
	}
	if addr.IsLoopback() {
		desc = append(desc, "loopback")
	}
	if addr.IsPrivate() {
		desc = append(desc, "private")
	}
	if addr.IsMulticast() {
		desc = append(desc, "multicast")
	}
	if addr.IsLinkLocal() {
		desc = append(desc, "link-local")
	}

	if addr.IsIPv4() {
		class := addr.GetIPv4Class()
		resp.IPv4Class = &class
		desc = append(desc, "Class "+string(class))
	} else if addr.IsIPv6() {
		ipv6Type := addr.GetIPv6Type()
		resp.IPv6Type = &ipv6Type
		desc = append(desc, string(ipv6Type))
	}

	if len(desc) > 0 {
		resp.Description = strings.Join(desc, ", ")
	} else {
		resp.Description = "public unicast"
	}

	return resp
}

func main() {
	http.HandleFunc("/parse", handleParse)
	http.HandleFunc("/format", handleFormat)
	http.HandleFunc("/classify", handleClassify)

	log.Println("IP Parser Server starting on :8303")
	log.Println("Endpoints:")
	log.Println("  POST /parse   - Parse an IP address")
	log.Println("  POST /format  - Format an IP address")
	log.Println("  POST /classify - Classify an IP address type")

	if err := http.ListenAndServe(":8303", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
