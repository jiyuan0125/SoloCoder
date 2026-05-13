package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"inventory-alert/models"
	"inventory-alert/service"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}

func ProductsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		products, err := service.GetProducts()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, products)

	case http.MethodPost:
		var product models.Product
		err := json.NewDecoder(r.Body).Decode(&product)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		id, err := service.CreateProduct(&product)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		product.ID = id
		writeJSON(w, http.StatusCreated, product)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func ProductHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/products/")
	if idStr == "" {
		writeError(w, http.StatusBadRequest, "product id is required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		product, err := service.GetProduct(id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if product == nil {
			writeError(w, http.StatusNotFound, "product not found")
			return
		}
		writeJSON(w, http.StatusOK, product)

	case http.MethodPut:
		var product models.Product
		err := json.NewDecoder(r.Body).Decode(&product)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		product.ID = id

		err = service.UpdateProduct(&product)
		if err != nil {
			if err.Error() == "product not found" {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, product)

	case http.MethodDelete:
		err := service.DeleteProduct(id)
		if err != nil {
			if err.Error() == "product not found" {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func ProductAlertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/products/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	alert, err := service.CheckAndCreateAlert(id)
	if err != nil {
		if dupErr, ok := err.(*service.DuplicateAlertError); ok {
			msg := fmt.Sprintf("duplicate alert, last alert at %s", dupErr.LastAlertTime.Format("2006-01-02 15:04:05"))
			writeError(w, http.StatusConflict, msg)
			return
		}
		if err.Error() == "product not found" {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if alert == nil {
		writeJSON(w, http.StatusOK, map[string]string{"message": "no alert needed"})
		return
	}

	writeJSON(w, http.StatusCreated, alert)
}

func ProductPurchaseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/products/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	var req struct {
		Quantity int `json:"quantity"`
	}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = service.ProcessPurchase(id, req.Quantity)
	if err != nil {
		if err.Error() == "product not found" {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "purchase processed successfully"})
}

func AlertsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	level := r.URL.Query().Get("level")
	alerts, err := service.GetAlerts(level)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, alerts)
}

func AlertHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/alerts/")
	if idStr == "" {
		writeError(w, http.StatusBadRequest, "alert id is required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid alert id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		alert, err := service.GetAlert(id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if alert == nil {
			writeError(w, http.StatusNotFound, "alert not found")
			return
		}
		writeJSON(w, http.StatusOK, alert)

	case http.MethodPut:
		var req struct {
			Status models.AlertStatus `json:"status"`
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		err = service.UpdateAlertStatus(id, req.Status)
		if err != nil {
			if err.Error() == "alert not found" {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		alert, _ := service.GetAlert(id)
		writeJSON(w, http.StatusOK, alert)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func AlertAssignHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/alerts/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid alert id")
		return
	}

	var req struct {
		AssignedTo string `json:"assigned_to"`
	}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = service.AssignAlert(id, req.AssignedTo)
	if err != nil {
		if err.Error() == "alert not found" {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	alert, _ := service.GetAlert(id)
	writeJSON(w, http.StatusOK, alert)
}

func StatisticsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	stats, err := service.GetCategoryStatistics()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, stats)
}
