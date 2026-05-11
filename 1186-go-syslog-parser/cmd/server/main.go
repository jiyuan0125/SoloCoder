package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"syslog-parser/pkg/api"
	"syslog-parser/pkg/syslog"
)

func main() {
	port := flag.String("port", "8200", "Server port")
	flag.Parse()

	if envPort := os.Getenv("SYSLOG_SERVER_PORT"); envPort != "" {
		*port = envPort
	}

	http.HandleFunc("/parse", handleParse)
	http.HandleFunc("/generate", handleGenerate)
	http.HandleFunc("/batch-parse", handleBatchParse)
	http.HandleFunc("/facility-severity", handleFacilitySeverity)

	log.Printf("Server starting on port %s", *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req api.ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(api.ParseResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	msg, err := syslog.ParseMessage(req.Message)
	if err != nil {
		json.NewEncoder(w).Encode(api.ParseResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse message: %v", err),
		})
		return
	}

	json.NewEncoder(w).Encode(api.ParseResponse{
		Success: true,
		Message: convertToAPIMessage(msg),
	})
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req api.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(api.GenerateResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	sd := convertFromAPISD(req.StructuredData)

	var msg *syslog.Message
	if req.Version == 0 {
		req.Version = 1
	}

	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}

	msg = &syslog.Message{
		PRI:             syslog.CalculatePRI(syslog.Facility(req.Facility), syslog.Severity(req.Severity)),
		Facility:        syslog.Facility(req.Facility),
		Severity:        syslog.Severity(req.Severity),
		Version:         req.Version,
		Timestamp:       req.Timestamp,
		TimestampFormat: "rfc3339",
		Hostname:        req.Hostname,
		AppName:         req.AppName,
		ProcID:          req.ProcID,
		MsgID:           req.MsgID,
		StructuredData:  sd,
		Message:         req.Message,
	}

	json.NewEncoder(w).Encode(api.GenerateResponse{
		Success: true,
		Message: msg.String(),
	})
}

func handleBatchParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req api.BatchParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(api.BatchParseResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	results := make([]*api.ParsedSyslogMsg, 0, len(req.Messages))
	for _, raw := range req.Messages {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}

		result := &api.ParsedSyslogMsg{Raw: raw}
		msg, err := syslog.ParseMessage(trimmed)
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Parsed = convertToAPIMessage(msg)
		}
		results = append(results, result)
	}

	json.NewEncoder(w).Encode(api.BatchParseResponse{
		Success:  true,
		Messages: results,
	})
}

func handleFacilitySeverity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	facilities := make([]api.FacilityInfo, 0, len(syslog.FacilityNames))
	for code, name := range syslog.FacilityNames {
		facilities = append(facilities, api.FacilityInfo{
			Code: int(code),
			Name: name,
		})
	}

	severities := make([]api.SeverityInfo, 0, len(syslog.SeverityNames))
	for code, name := range syslog.SeverityNames {
		severities = append(severities, api.SeverityInfo{
			Code: int(code),
			Name: name,
		})
	}

	json.NewEncoder(w).Encode(api.FacilitySeverityResponse{
		Success:  true,
		Facility: facilities,
		Severity: severities,
	})
}

func convertToAPIMessage(msg *syslog.Message) *api.SyslogMsg {
	apiMsg := &api.SyslogMsg{
		PRI:             msg.PRI,
		Facility:        int(msg.Facility),
		FacilityName:    syslog.FacilityNames[msg.Facility],
		Severity:        int(msg.Severity),
		SeverityName:    syslog.SeverityNames[msg.Severity],
		Version:         msg.Version,
		Timestamp:       msg.Timestamp,
		TimestampFormat: msg.TimestampFormat,
		Hostname:        msg.Hostname,
		AppName:         msg.AppName,
		ProcID:          msg.ProcID,
		MsgID:           msg.MsgID,
		Message:         msg.Message,
	}

	if msg.StructuredData != nil {
		sdList := make([]api.SDElement, 0, len(msg.StructuredData))
		for _, elem := range msg.StructuredData {
			apiElem := api.SDElement{
				ID:     elem.ID,
				Params: make([]api.SDParam, 0, len(elem.Params)),
			}
			for _, p := range elem.Params {
				apiElem.Params = append(apiElem.Params, api.SDParam{
					Name:  p.Name,
					Value: p.Value,
				})
			}
			sdList = append(sdList, apiElem)
		}
		apiMsg.StructuredData = sdList
	}

	return apiMsg
}

func convertFromAPISD(apiSD []api.SDElement) syslog.StructuredData {
	if len(apiSD) == 0 {
		return nil
	}

	sd := make(syslog.StructuredData, 0, len(apiSD))
	for _, apiElem := range apiSD {
		elem := &syslog.SDElement{
			ID:     apiElem.ID,
			Params: make([]syslog.SDParam, 0, len(apiElem.Params)),
		}
		for _, apiParam := range apiElem.Params {
			elem.Params = append(elem.Params, syslog.SDParam{
				Name:  apiParam.Name,
				Value: apiParam.Value,
			})
		}
		sd = append(sd, elem)
	}

	return sd
}
