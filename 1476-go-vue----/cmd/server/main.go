package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gasstation/internal/common"
	"gasstation/internal/core"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	store   *core.Store
	service *core.RefuelService
}

func NewServer() *Server {
	store := core.NewStore()
	return &Server{
		store:   store,
		service: core.NewRefuelService(store),
	}
}

func writeJSON(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	resp := common.APIResponse{
		Code:    code,
		Message: message,
		Data:    data,
	}
	json.NewEncoder(w).Encode(resp)
}

func readJSON(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.Unmarshal(body, v)
}

func (s *Server) createStation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}
	var req common.CreateStationRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	station := &common.GasStation{
		Name:      req.Name,
		Address:   req.Address,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Phone:     req.Phone,
		IsOpen:    req.IsOpen,
	}
	if err := s.store.CreateStation(station); err != nil {
		writeJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusCreated, "station created", station)
}

func (s *Server) listStations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}
	stations := s.store.ListStations()
	writeJSON(w, http.StatusOK, "success", stations)
}

func (s *Server) getStation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/stations/")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, "station id required", nil)
		return
	}
	station, err := s.store.GetStation(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, "success", station)
}

func (s *Server) updateStation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		writeJSON(w, http.StatusBadRequest, "station id required", nil)
		return
	}
	id := pathParts[1]
	var req common.UpdateStationRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	station, err := s.store.UpdateStation(id, req)
	if err != nil {
		writeJSON(w, http.StatusNotFound, err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, "station updated", station)
}

func (s *Server) addFuelToStation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		writeJSON(w, http.StatusBadRequest, "station id required", nil)
		return
	}
	stationID := pathParts[1]
	var body struct {
		FuelCode string `json:"fuel_code"`
		FuelName string `json:"fuel_name"`
		Price    int64  `json:"price"`
		Stock    int64  `json:"stock"`
	}
	if err := readJSON(r, &body); err != nil {
		writeJSON(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	fuel, err := s.store.AddFuelToStation(stationID, body.FuelCode, body.FuelName, body.Price, body.Stock)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusCreated, "fuel added", fuel)
}

func (s *Server) listStationFuels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		writeJSON(w, http.StatusBadRequest, "station id required", nil)
		return
	}
	stationID := pathParts[1]
	fuels, err := s.store.ListStationFuels(stationID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, "success", fuels)
}

func (s *Server) registerMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}
	var req common.RegisterMemberRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	member, err := s.store.RegisterMember(req.Phone)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusCreated, "member registered", member)
}

func (s *Server) listMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}
	members := s.store.ListMembers()
	writeJSON(w, http.StatusOK, "success", members)
}

func (s *Server) getMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/members/")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, "member id required", nil)
		return
	}
	member, err := s.store.GetMember(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, "success", member)
}

func (s *Server) getMemberByPhone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}
	phone := r.URL.Query().Get("phone")
	if phone == "" {
		writeJSON(w, http.StatusBadRequest, "phone query parameter required", nil)
		return
	}
	member, err := s.store.GetMemberByPhone(phone)
	if err != nil {
		writeJSON(w, http.StatusNotFound, err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, "success", member)
}

func (s *Server) refuel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}
	var req common.RefuelRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	record, err := s.service.ProcessRefuel(req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusCreated, "refuel processed", record)
}

func (s *Server) listRecords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}
	stationID := r.URL.Query().Get("station_id")
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	start := time.Now().AddDate(0, -1, 0)
	end := time.Now()

	if startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			start = t
		}
	}
	if endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			end = t
		}
	}

	records := s.store.ListRecords(stationID, start, end)
	writeJSON(w, http.StatusOK, "success", records)
}

func (s *Server) exportRecords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}
	stationID := r.URL.Query().Get("station_id")
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	start := time.Now().AddDate(0, -1, 0)
	end := time.Now()

	if startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			start = t
		}
	}
	if endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			end = t
		}
	}

	csv := s.store.ExportRecordsCSV(stationID, start, end)
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=records.csv")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(csv))
}

func (s *Server) submitPriceChange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}
	var req common.SubmitPriceChangeRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	pr, err := s.store.SubmitPriceChangeRequest(req.StationID, req.FuelCode, req.NewPrice, "operator")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusCreated, "price change request submitted", pr)
}

func (s *Server) listPriceRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}
	stationID := r.URL.Query().Get("station_id")
	reqs := s.store.ListPriceChangeRequests(stationID)
	writeJSON(w, http.StatusOK, "success", reqs)
}

func (s *Server) reviewPriceChange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/price-requests/")
	if id == "" || strings.Contains(id, "/") {
		writeJSON(w, http.StatusBadRequest, "price request id required", nil)
		return
	}
	var req common.ReviewPriceChangeRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	if err := s.store.ReviewPriceChangeRequest(id, req.Approved, req.ReviewedBy); err != nil {
		writeJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, "price request reviewed", nil)
}

func (s *Server) getPriceHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}
	stationID := r.URL.Query().Get("station_id")
	fuelCode := r.URL.Query().Get("fuel_code")
	if stationID == "" || fuelCode == "" {
		writeJSON(w, http.StatusBadRequest, "station_id and fuel_code required", nil)
		return
	}
	history := s.store.GetPriceHistory(stationID, fuelCode)
	writeJSON(w, http.StatusOK, "success", history)
}

func (s *Server) listRestockTodos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}
	todos := s.store.ListRestockTodos()
	writeJSON(w, http.StatusOK, "success", todos)
}

func (s *Server) completeRestockTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/restock-todos/")
	if strings.Contains(id, "/complete") {
		id = strings.TrimSuffix(id, "/complete")
	}
	if id == "" {
		writeJSON(w, http.StatusBadRequest, "restock todo id required", nil)
		return
	}
	var body struct {
		NewStock int64 `json:"new_stock"`
	}
	if err := readJSON(r, &body); err != nil {
		writeJSON(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	if err := s.store.CompleteRestockTodo(id, body.NewStock); err != nil {
		writeJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, "restock todo completed", nil)
}

func (s *Server) getPort() string {
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	flagPort := flag.String("port", "", "server port")
	flag.Parse()
	if *flagPort != "" {
		port = *flagPort
	}

	if _, err := strconv.Atoi(port); err != nil {
		port = "8080"
	}
	return ":" + port
}

func main() {
	server := NewServer()
	port := server.getPort()

	mux := http.NewServeMux()

	mux.HandleFunc("/stations", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			server.createStation(w, r)
		} else {
			server.listStations(w, r)
		}
	})
	mux.HandleFunc("/stations/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.Trim(r.URL.Path, "/")
		parts := strings.Split(path, "/")
		if len(parts) == 2 {
			if r.Method == http.MethodGet {
				server.getStation(w, r)
			} else if r.Method == http.MethodPut {
				server.updateStation(w, r)
			} else {
				writeJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
			}
		} else if len(parts) == 3 && parts[2] == "fuels" {
			if r.Method == http.MethodPost {
				server.addFuelToStation(w, r)
			} else if r.Method == http.MethodGet {
				server.listStationFuels(w, r)
			} else {
				writeJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
			}
		} else {
			writeJSON(w, http.StatusNotFound, "not found", nil)
		}
	})

	mux.HandleFunc("/members", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			server.registerMember(w, r)
		} else {
			server.listMembers(w, r)
		}
	})
	mux.HandleFunc("/members/", func(w http.ResponseWriter, r *http.Request) {
		server.getMember(w, r)
	})
	mux.HandleFunc("/members/by-phone", server.getMemberByPhone)

	mux.HandleFunc("/refuel", server.refuel)
	mux.HandleFunc("/records", server.listRecords)
	mux.HandleFunc("/records/export", server.exportRecords)

	mux.HandleFunc("/price-requests", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			server.submitPriceChange(w, r)
		} else {
			server.listPriceRequests(w, r)
		}
	})
	mux.HandleFunc("/price-requests/", server.reviewPriceChange)
	mux.HandleFunc("/price-history", server.getPriceHistory)

	mux.HandleFunc("/restock-todos", server.listRestockTodos)
	mux.HandleFunc("/restock-todos/", server.completeRestockTodo)

	fmt.Printf("Server starting on %s\n", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
