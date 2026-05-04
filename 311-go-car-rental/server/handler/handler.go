package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"carrental/common"
	"carrental/server/service"
)

type Handler struct {
	carService   *service.CarService
	orderService *service.OrderService
}

func NewHandler(carService *service.CarService, orderService *service.OrderService) *Handler {
	return &Handler{
		carService:   carService,
		orderService: orderService,
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func readBody(r io.ReadCloser, v interface{}) error {
	defer r.Close()
	return json.NewDecoder(r).Decode(v)
}

func (h *Handler) HandleCars(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.handleCreateCar(w, r)
	case http.MethodGet:
		h.handleListCars(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
	}
}

func (h *Handler) handleCreateCar(w http.ResponseWriter, r *http.Request) {
	var req common.CreateCarRequest
	if err := readBody(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	car, err := h.carService.CreateCar(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, common.CreateCarResponse{Car: car})
}

func (h *Handler) handleListCars(w http.ResponseWriter, r *http.Request) {
	cars, err := h.carService.ListCars()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, common.ListCarsResponse{Cars: cars})
}

func (h *Handler) HandleUpdateCarStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
		return
	}

	var req common.UpdateCarStatusRequest
	if err := readBody(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	car, err := h.carService.UpdateCarStatus(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, common.UpdateCarStatusResponse{Car: car})
}

func (h *Handler) HandleCheckAvailability(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
		return
	}

	carID := r.URL.Query().Get("car_id")
	pickupDateStr := r.URL.Query().Get("pickup_date")
	returnDateStr := r.URL.Query().Get("return_date")

	if carID == "" || pickupDateStr == "" || returnDateStr == "" {
		writeError(w, http.StatusBadRequest, http.ErrMissingFile)
		return
	}

	pickupDate, err := common.ParseDate(pickupDateStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	returnDate, err := common.ParseDate(returnDateStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	available, err := h.orderService.CheckAvailability(carID, pickupDate, returnDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, common.CheckAvailabilityResponse{Available: available})
}

func (h *Handler) HandleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.handleCreateOrder(w, r)
	case http.MethodGet:
		h.handleListOrders(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
	}
}

func (h *Handler) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	var req common.CreateOrderRequest
	if err := readBody(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	order, err := h.orderService.CreateOrder(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, common.CreateOrderResponse{Order: order})
}

func (h *Handler) handleListOrders(w http.ResponseWriter, r *http.Request) {
	userPhone := r.URL.Query().Get("user_phone")
	orderID := r.URL.Query().Get("order_id")

	if orderID != "" {
		order, err := h.orderService.GetOrder(orderID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if order == nil {
			writeError(w, http.StatusNotFound, service.ErrOrderNotFound)
			return
		}
		writeJSON(w, http.StatusOK, common.GetOrderResponse{Order: order})
		return
	}

	var orders []*common.Order
	var err error

	if userPhone != "" {
		orders, err = h.orderService.ListOrdersByPhone(userPhone)
	} else {
		orders, err = h.orderService.ListOrders()
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, common.ListOrdersResponse{Orders: orders})
}

func (h *Handler) HandleReturnCar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
		return
	}

	var req common.ReturnCarRequest
	if err := readBody(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	order, err := h.orderService.ReturnCar(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, common.ReturnCarResponse{Order: order})
}
