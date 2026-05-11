package main

import (
	"encoding/json"
	"net/http"
	"taxisystem/common"
	"taxisystem/core"

	"github.com/gorilla/mux"
)

type Server struct {
	vm *core.VehicleManager
	fc *core.FareCalculator
	dm *core.DispatchManager
	rm *core.RatingManager
	mm *core.MetricsManager
}

func NewServer() *Server {
	vm := core.NewVehicleManager()
	fc := core.NewFareCalculator()
	dm := core.NewDispatchManager(vm, fc)
	rm := core.NewRatingManager(dm)
	mm := core.NewMetricsManager(dm)

	return &Server{
		vm: vm,
		fc: fc,
		dm: dm,
		rm: rm,
		mm: mm,
	}
}

func writeResponse(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := common.APIResponse{
		Code:    code,
		Message: message,
		Data:    data,
	}

	json.NewEncoder(w).Encode(resp)
}

func (s *Server) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	var req common.CreateVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeResponse(w, 400, "参数解析失败: "+err.Error(), nil)
		return
	}

	vehicle, err := s.vm.AddVehicle(req.PlateNumber, req.DriverName, req.DriverPhone, req.Latitude, req.Longitude)
	if err != nil {
		writeResponse(w, 400, err.Error(), nil)
		return
	}

	writeResponse(w, 0, "成功", vehicle)
}

func (s *Server) ListVehicles(w http.ResponseWriter, r *http.Request) {
	vehicles := s.vm.ListVehicles()
	writeResponse(w, 0, "成功", vehicles)
}

func (s *Server) GetVehicle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	plate := vars["plate"]

	vehicle, err := s.vm.GetVehicle(plate)
	if err != nil {
		writeResponse(w, 404, err.Error(), nil)
		return
	}

	writeResponse(w, 0, "成功", vehicle)
}

func (s *Server) UpdateVehicleStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	plate := vars["plate"]

	var req common.UpdateVehicleStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeResponse(w, 400, "参数解析失败: "+err.Error(), nil)
		return
	}

	err := s.vm.UpdateStatus(plate, req.Status)
	if err != nil {
		writeResponse(w, 400, err.Error(), nil)
		return
	}

	writeResponse(w, 0, "成功", nil)
}

func (s *Server) UpdateVehicleLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	plate := vars["plate"]

	var req common.UpdateVehicleLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeResponse(w, 400, "参数解析失败: "+err.Error(), nil)
		return
	}

	err := s.vm.UpdateLocation(plate, req.Latitude, req.Longitude)
	if err != nil {
		writeResponse(w, 400, err.Error(), nil)
		return
	}

	writeResponse(w, 0, "成功", nil)
}

func (s *Server) EstimateFare(w http.ResponseWriter, r *http.Request) {
	var req common.EstimateFareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeResponse(w, 400, "参数解析失败: "+err.Error(), nil)
		return
	}

	fare, distance, duration := s.dm.EstimateFare(req.PickupLocation, req.DestLocation)

	result := map[string]interface{}{
		"estimated_fare":      fare,
		"estimated_distance":  distance,
		"estimated_duration":  duration,
	}

	writeResponse(w, 0, "成功", result)
}

func (s *Server) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req common.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeResponse(w, 400, "参数解析失败: "+err.Error(), nil)
		return
	}

	order, err := s.dm.CreateOrder(req.PassengerID, req.PassengerPhone, req.PickupLocation, req.DestLocation)
	if err != nil {
		writeResponse(w, 400, err.Error(), nil)
		return
	}

	writeResponse(w, 0, "成功", order)
}

func (s *Server) ListOrders(w http.ResponseWriter, r *http.Request) {
	orders := s.dm.ListOrders()
	writeResponse(w, 0, "成功", orders)
}

func (s *Server) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]

	order, err := s.dm.GetOrder(orderID)
	if err != nil {
		writeResponse(w, 404, err.Error(), nil)
		return
	}

	writeResponse(w, 0, "成功", order)
}

func (s *Server) AcceptOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]

	var req common.AcceptOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeResponse(w, 400, "参数解析失败: "+err.Error(), nil)
		return
	}

	err := s.dm.AcceptOrder(orderID, req.PlateNumber)
	if err != nil {
		writeResponse(w, 400, err.Error(), nil)
		return
	}

	writeResponse(w, 0, "接单成功", nil)
}

func (s *Server) StartTrip(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]

	var req common.StartTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeResponse(w, 400, "参数解析失败: "+err.Error(), nil)
		return
	}

	err := s.dm.StartTrip(orderID, req.ActualDistance, req.ActualDuration, req.LowSpeedMinutes)
	if err != nil {
		writeResponse(w, 400, err.Error(), nil)
		return
	}

	writeResponse(w, 0, "行程开始", nil)
}

func (s *Server) CompleteTrip(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]

	order, err := s.dm.CompleteTrip(orderID)
	if err != nil {
		writeResponse(w, 400, err.Error(), nil)
		return
	}

	writeResponse(w, 0, "行程结束", order)
}

func (s *Server) SubmitRating(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]

	var req common.SubmitRatingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeResponse(w, 400, "参数解析失败: "+err.Error(), nil)
		return
	}

	rating, err := s.rm.SubmitRating(orderID, req.Stars, req.Content)
	if err != nil {
		writeResponse(w, 400, err.Error(), nil)
		return
	}

	writeResponse(w, 0, "评价成功", rating)
}

func (s *Server) ListComplaints(w http.ResponseWriter, r *http.Request) {
	complaints := s.rm.GetPendingComplaints()
	writeResponse(w, 0, "成功", complaints)
}

func (s *Server) HandleComplaint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	complaintID := vars["id"]

	err := s.rm.HandleComplaint(complaintID)
	if err != nil {
		writeResponse(w, 404, err.Error(), nil)
		return
	}

	writeResponse(w, 0, "投诉已处理", nil)
}

func (s *Server) GetMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := s.mm.GetMetrics()
	writeResponse(w, 0, "成功", metrics)
}
