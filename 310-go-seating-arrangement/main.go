package main

import (
	"encoding/json"
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	LoadData()
	StartAutoSave()

	setupRoutes()

	go func() {
		fmt.Println("Server starting on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
	StopAutoSave()
	SaveData()
	fmt.Println("Data saved. Goodbye.")
}

func setupRoutes() {
	http.HandleFunc("/api/venues", venuesHandler)
	http.HandleFunc("/api/venues/", venueHandler)

	http.HandleFunc("/api/events", eventsHandler)
	http.HandleFunc("/api/events/", eventHandler)

	http.HandleFunc("/api/orders", ordersHandler)
	http.HandleFunc("/api/orders/", orderHandler)
}

func venuesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createVenueHandler(w, r)
	case http.MethodGet:
		listVenuesHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func venueHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/venues/")
	if path == "" || strings.Contains(path, "/") {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	venue, err := GetVenue(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	jsonResponse(w, venue)
}

func createVenueHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateVenueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	venue, err := CreateVenue(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	go SaveData()
	jsonResponse(w, venue)
}

func listVenuesHandler(w http.ResponseWriter, r *http.Request) {
	venues := GetAllVenues()
	jsonResponse(w, venues)
}

func eventsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createEventHandler(w, r)
	case http.MethodGet:
		listEventsHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func eventHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/events/")
	parts := strings.Split(path, "/")

	if len(parts) < 1 || parts[0] == "" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	eventID := parts[0]

	if len(parts) == 1 {
		if r.Method == http.MethodGet {
			event, err := GetEvent(eventID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			jsonResponse(w, event)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if len(parts) >= 2 {
		switch parts[1] {
		case "stats":
			if r.Method == http.MethodGet {
				getEventStatsHandler(w, r, eventID)
				return
			}
		case "export":
			if r.Method == http.MethodGet {
				exportEventDetailsHandler(w, r, eventID)
				return
			}
		case "sections":
			if len(parts) >= 4 && parts[3] == "seats" && r.Method == http.MethodGet {
				getUserViewableSeatsHandler(w, r, eventID, parts[2])
				return
			}
		case "seats":
			if len(parts) == 3 && parts[2] == "release" && r.Method == http.MethodPost {
				adminReleaseSeatHandler(w, r, eventID)
				return
			}
		}
	}

	http.Error(w, "Not found", http.StatusNotFound)
}

func createEventHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	event, err := CreateEvent(req)
	if err != nil {
		if err == ErrInvalidEventName || err == ErrVenueNotFound {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go SaveData()
	jsonResponse(w, event)
}

func listEventsHandler(w http.ResponseWriter, r *http.Request) {
	events := GetAllEvents()
	jsonResponse(w, events)
}

func getEventStatsHandler(w http.ResponseWriter, r *http.Request, eventID string) {
	stats, err := GetEventSeatStats(eventID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, stats)
}

func getUserViewableSeatsHandler(w http.ResponseWriter, r *http.Request, eventID string, sectionName string) {
	seats, err := GetUserViewableSeats(eventID, sectionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, seats)
}

func exportEventDetailsHandler(w http.ResponseWriter, r *http.Request, eventID string) {
	details, err := GetSeatedDetails(eventID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	eventOrders := GetOrdersByEvent(eventID)
	orderMap := make(map[string]*Order)
	for _, o := range eventOrders {
		orderMap[o.SeatKey] = o
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=seating-%s.csv", eventID))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	writer.Write([]string{"区域", "排号", "座位索引", "状态", "订单ID", "订单状态", "创建时间", "支付时间"})

	for _, seat := range details {
		order := orderMap[seat.SeatKey]
		orderID := ""
		orderStatus := ""
		createdAt := ""
		paidAt := ""

		if order != nil {
			orderID = order.ID
			orderStatus = order.Status
			createdAt = order.CreatedAt.Format(time.RFC3339)
			if !order.PaidAt.IsZero() {
				paidAt = order.PaidAt.Format(time.RFC3339)
			}
		}

		writer.Write([]string{
			seat.Section,
			fmt.Sprintf("%d", seat.RowNumber),
			fmt.Sprintf("%d", seat.SeatIndex),
			seat.Status,
			orderID,
			orderStatus,
			createdAt,
			paidAt,
		})
	}
}

func adminReleaseSeatHandler(w http.ResponseWriter, r *http.Request, eventID string) {
	var req struct {
		SeatKey string `json:"seat_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.SeatKey == "" {
		http.Error(w, "seat_key is required", http.StatusBadRequest)
		return
	}

	err := AdminReleaseSeat(eventID, req.SeatKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	go SaveData()
	jsonResponse(w, map[string]string{"message": "座位已释放"})
}

func ordersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createOrderHandler(w, r)
	case http.MethodGet:
		listOrdersHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func orderHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/orders/")
	parts := strings.Split(path, "/")

	if len(parts) < 1 || parts[0] == "" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	orderID := parts[0]

	if len(parts) == 1 {
		if r.Method == http.MethodGet {
			order, err := GetOrder(orderID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			jsonResponse(w, order)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if len(parts) == 2 && r.Method == http.MethodPost {
		switch parts[1] {
		case "pay":
			payOrderHandler(w, r, orderID)
			return
		case "cancel":
			cancelOrderHandler(w, r, orderID)
			return
		}
	}

	http.Error(w, "Not found", http.StatusNotFound)
}

func createOrderHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order, err := CreateOrder(req)
	if err != nil {
		if err == ErrSeatAlreadyLocked || err == ErrSeatAlreadySold {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	go SaveData()
	jsonResponse(w, order)
}

func listOrdersHandler(w http.ResponseWriter, r *http.Request) {
	orders := GetAllOrders()
	jsonResponse(w, orders)
}

func payOrderHandler(w http.ResponseWriter, r *http.Request, orderID string) {
	order, err := PayOrder(orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	go SaveData()
	jsonResponse(w, order)
}

func cancelOrderHandler(w http.ResponseWriter, r *http.Request, orderID string) {
	order, err := CancelOrder(orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	go SaveData()
	jsonResponse(w, order)
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
