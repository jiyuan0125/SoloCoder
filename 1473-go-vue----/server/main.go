package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"transit-system/common"
	"transit-system/core"
)

var transitSystem *core.TransitSystem

func main() {
	var port string
	flag.StringVar(&port, "port", "", "HTTP server port (default: 8080)")
	flag.Parse()

	if port == "" {
		port = os.Getenv("PORT")
		if port == "" {
			port = "8902"
		}
	}

	transitSystem = core.NewTransitSystem()
	initializeSampleData()
	transitSystem.Start()

	http.HandleFunc("/api/stations", handleStations)
	http.HandleFunc("/api/stations/", handleStationByCode)
	http.HandleFunc("/api/lines", handleLines)
	http.HandleFunc("/api/lines/", handleLineByCode)
	http.HandleFunc("/api/vehicles", handleVehicles)
	http.HandleFunc("/api/vehicles/", handleVehicleByID)
	http.HandleFunc("/api/lines/", handleLineVehicles)

	fmt.Printf("Server starting on port %s...\n", port)
	server := &http.Server{Addr: ":" + port}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		fmt.Println("\nShutting down server...")
		transitSystem.Stop()
		os.Exit(0)
	}()

	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("Server error: %v\n", err)
		transitSystem.Stop()
		os.Exit(1)
	}
}

func initializeSampleData() {
	stations := []struct {
		Code      string
		Name      string
		Latitude  float64
		Longitude float64
	}{
		{"S001", "北京站", 39.9042, 116.4074},
		{"S002", "建国门", 39.9075, 116.4280},
		{"S003", "朝阳门", 39.9243, 116.4378},
		{"S004", "东四", 39.9243, 116.4170},
		{"S005", "北新桥", 39.9431, 116.4170},
		{"S006", "雍和宫", 39.9530, 116.4170},
		{"S007", "安定门", 39.9530, 116.4074},
		{"S008", "鼓楼大街", 39.9431, 116.4074},
		{"S009", "西直门", 39.9431, 116.3568},
		{"S010", "车公庄", 39.9332, 116.3568},
		{"S011", "阜成门", 39.9243, 116.3568},
		{"S012", "复兴门", 39.9075, 116.3568},
	}

	for _, s := range stations {
		_, _ = transitSystem.AddStation(s.Code, s.Name, s.Latitude, s.Longitude)
	}

	line1UpStations := []string{"S001", "S002", "S003", "S004", "S005", "S006", "S007", "S008"}
	line1DownStations := []string{"S008", "S007", "S006", "S005", "S004", "S003", "S002", "S001"}
	line2UpStations := []string{"S009", "S010", "S011", "S012", "S002", "S006", "S007", "S009"}

	_, _ = transitSystem.AddLine("L1", "1号线", core.DirectionUp, "05:00", "23:00", line1UpStations)
	_, _ = transitSystem.AddLine("L1", "1号线", core.DirectionDown, "05:10", "23:10", line1DownStations)
	_, _ = transitSystem.AddLine("L2", "2号线", core.DirectionUp, "05:30", "22:30", line2UpStations)

	_, _ = transitSystem.AddVehicle("V001", "L1", core.DirectionUp)
	_, _ = transitSystem.AddVehicle("V002", "L1", core.DirectionUp)
	_, _ = transitSystem.AddVehicle("V003", "L1", core.DirectionDown)
	_, _ = transitSystem.AddVehicle("V004", "L2", core.DirectionUp)

	transitSystem.RefreshStationTypes()
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func handleStations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		stations := transitSystem.GetAllStations()
		stationDTOs := make([]common.StationDTO, 0, len(stations))
		for _, station := range stations {
			stationDTOs = append(stationDTOs, stationToDTO(station))
		}
		writeJSON(w, http.StatusOK, common.SuccessResponse(stationDTOs))
	case http.MethodPost:
		var req common.CreateStationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, common.ErrorResponseMessage("invalid request body"))
			return
		}

		station, err := transitSystem.AddStation(req.Code, req.Name, req.Latitude, req.Longitude)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, common.ErrorResponseMessage(err.Error()))
			return
		}
		writeJSON(w, http.StatusCreated, common.SuccessResponse(stationToDTO(station)))
	default:
		writeJSON(w, http.StatusMethodNotAllowed, common.ErrorResponseMessage("method not allowed"))
	}
}

func handleStationByCode(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/stations/")
	code := path
	if code == "" {
		writeJSON(w, http.StatusNotFound, common.ErrorResponseMessage("station code not found"))
		return
	}

	station, ok := transitSystem.GetStation(code)
	if !ok {
		writeJSON(w, http.StatusNotFound, common.ErrorResponseMessage("station not found"))
		return
	}
	writeJSON(w, http.StatusOK, common.SuccessResponse(stationToDTO(station)))
}

func handleLines(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		lines := transitSystem.GetAllLines()
		lineDTOs := make([]common.LineDTO, 0, len(lines))
		for _, line := range lines {
			lineDTOs = append(lineDTOs, lineToDTO(line))
		}
		writeJSON(w, http.StatusOK, common.SuccessResponse(lineDTOs))
	case http.MethodPost:
		var req common.CreateLineRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, common.ErrorResponseMessage("invalid request body"))
			return
		}

		direction := core.DirectionUp
		if strings.EqualFold(req.Direction, "down") || strings.EqualFold(req.Direction, "下行") {
			direction = core.DirectionDown
		}

		line, err := transitSystem.AddLine(req.Code, req.Name, direction, req.FirstTime, req.LastTime, req.StationCodes)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, common.ErrorResponseMessage(err.Error()))
			return
		}
		transitSystem.RefreshStationTypes()
		writeJSON(w, http.StatusCreated, common.SuccessResponse(lineToDTO(line)))
	default:
		writeJSON(w, http.StatusMethodNotAllowed, common.ErrorResponseMessage("method not allowed"))
	}
}

func handleLineByCode(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/lines/")
	parts := strings.SplitN(path, "/", 2)

	if len(parts) == 0 {
		writeJSON(w, http.StatusNotFound, common.ErrorResponseMessage("line code not found"))
		return
	}

	code := parts[0]
	directionStr := "up"
	if len(parts) == 2 && parts[1] != "" {
		directionStr = parts[1]
	}

	direction := core.DirectionUp
	if strings.EqualFold(directionStr, "down") || strings.EqualFold(directionStr, "下行") {
		direction = core.DirectionDown
	}

	switch r.Method {
	case http.MethodGet:
		line, ok := transitSystem.GetLine(code, direction)
		if !ok {
			writeJSON(w, http.StatusNotFound, common.ErrorResponseMessage("line not found"))
			return
		}
		writeJSON(w, http.StatusOK, common.SuccessResponse(lineToDTO(line)))
	case http.MethodPut:
		var req common.UpdateLineRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, common.ErrorResponseMessage("invalid request body"))
			return
		}

		updates := make(map[string]interface{})
		if req.Name != nil {
			updates["name"] = *req.Name
		}
		if req.FirstTime != nil {
			updates["first_time"] = *req.FirstTime
		}
		if req.LastTime != nil {
			updates["last_time"] = *req.LastTime
		}
		if req.StationCodes != nil {
			updates["station_codes"] = req.StationCodes
		}

		if err := transitSystem.UpdateLine(code, direction, updates); err != nil {
			writeJSON(w, http.StatusBadRequest, common.ErrorResponseMessage(err.Error()))
			return
		}
		transitSystem.RefreshStationTypes()

		line, _ := transitSystem.GetLine(code, direction)
		writeJSON(w, http.StatusOK, common.SuccessResponse(lineToDTO(line)))
	default:
		writeJSON(w, http.StatusMethodNotAllowed, common.ErrorResponseMessage("method not allowed"))
	}
}

func handleLineVehicles(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.URL.Path, "/vehicles") {
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/lines/")
	path = strings.TrimSuffix(path, "/vehicles")
	parts := strings.SplitN(path, "/", 2)

	if len(parts) == 0 {
		writeJSON(w, http.StatusNotFound, common.ErrorResponseMessage("line code not found"))
		return
	}

	code := parts[0]
	directionStr := "up"
	if len(parts) == 2 && parts[1] != "" {
		directionStr = parts[1]
	}

	direction := core.DirectionUp
	if strings.EqualFold(directionStr, "down") || strings.EqualFold(directionStr, "下行") {
		direction = core.DirectionDown
	}

	vehicles := transitSystem.GetVehiclesByLine(code, direction)
	vehicleDTOs := make([]common.VehicleDTO, 0, len(vehicles))
	for _, vehicle := range vehicles {
		vehicleDTOs = append(vehicleDTOs, vehicleToDTO(vehicle))
	}
	writeJSON(w, http.StatusOK, common.SuccessResponse(vehicleDTOs))
}

func handleVehicles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		vehicles := transitSystem.GetAllVehicles()
		vehicleDTOs := make([]common.VehicleDTO, 0, len(vehicles))
		for _, vehicle := range vehicles {
			vehicleDTOs = append(vehicleDTOs, vehicleToDTO(vehicle))
		}
		writeJSON(w, http.StatusOK, common.SuccessResponse(vehicleDTOs))
	case http.MethodPost:
		var req common.CreateVehicleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, common.ErrorResponseMessage("invalid request body"))
			return
		}

		direction := core.DirectionUp
		if strings.EqualFold(req.Direction, "down") || strings.EqualFold(req.Direction, "下行") {
			direction = core.DirectionDown
		}

		vehicle, err := transitSystem.AddVehicle(req.ID, req.LineCode, direction)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, common.ErrorResponseMessage(err.Error()))
			return
		}
		writeJSON(w, http.StatusCreated, common.SuccessResponse(vehicleToDTO(vehicle)))
	default:
		writeJSON(w, http.StatusMethodNotAllowed, common.ErrorResponseMessage("method not allowed"))
	}
}

func handleVehicleByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/vehicles/")
	id := path
	if id == "" {
		writeJSON(w, http.StatusNotFound, common.ErrorResponseMessage("vehicle ID not found"))
		return
	}

	vehicle, ok := transitSystem.GetVehicle(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, common.ErrorResponseMessage("vehicle not found"))
		return
	}
	writeJSON(w, http.StatusOK, common.SuccessResponse(vehicleToDTO(vehicle)))
}

func stationToDTO(station *core.Station) common.StationDTO {
	return common.StationDTO{
		Code:      station.Code,
		Name:      station.Name,
		Latitude:  station.Latitude,
		Longitude: station.Longitude,
		Type:      strconv.Itoa(int(station.Type)),
	}
}

func lineToDTO(line *core.Line) common.LineDTO {
	stationDTOs := make([]common.StationDTO, 0, len(line.Stations))
	for _, station := range line.Stations {
		stationDTOs = append(stationDTOs, stationToDTO(station))
	}

	return common.LineDTO{
		Code:          line.Code,
		Name:          line.Name,
		Direction:     line.Direction.String(),
		FirstTime:     core.FormatTime(line.FirstHour, line.FirstMinute),
		LastTime:      core.FormatTime(line.LastHour, line.LastMinute),
		Stations:      stationDTOs,
		IsOperational: line.IsOperational,
		HasError:      line.HasError,
		ErrorReason:   line.ErrorReason,
	}
}

func vehicleToDTO(vehicle *core.Vehicle) common.VehicleDTO {
	vehicle.Mutex.RLock()
	defer vehicle.Mutex.RUnlock()

	return common.VehicleDTO{
		ID:            vehicle.ID,
		LineCode:      vehicle.Line.Code,
		LineDirection: vehicle.Line.Direction.String(),
		CurrentStation: vehicle.CurrentStation,
		Status:        vehicle.Status.String(),
		Latitude:      vehicle.Latitude,
		Longitude:     vehicle.Longitude,
	}
}
