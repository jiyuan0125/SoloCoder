package handler

import (
	"billing-service/service"
	"encoding/json"
	"net/http"
	"strings"
)

type ConfigHandler struct {
	configService *service.ConfigService
}

func NewConfigHandler(configService *service.ConfigService) *ConfigHandler {
	return &ConfigHandler{configService: configService}
}

type SetConfigRequest struct {
	Value       string `json:"value"`
	Description string `json:"description"`
}

type SetPriceRequest struct {
	Price float64 `json:"price"`
}

func (h *ConfigHandler) HandleConfigs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.setConfig(w, r)
	default:
		RespondError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *ConfigHandler) HandleConfigByKey(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/configs/")
	key := path

	if key == "" {
		RespondError(w, http.StatusBadRequest, "invalid config key")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getConfig(w, r, key)
	case http.MethodPut:
		h.updateConfig(w, r, key)
	default:
		RespondError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *ConfigHandler) getConfig(w http.ResponseWriter, r *http.Request, key string) {
	if key == "sms-unit-price" {
		price, err := h.configService.GetSmsUnitPrice()
		if err != nil {
			RespondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		RespondSuccess(w, map[string]float64{"price": price})
		return
	}

	if key == "storage-unit-price" {
		price, err := h.configService.GetStorageUnitPrice()
		if err != nil {
			RespondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		RespondSuccess(w, map[string]float64{"price": price})
		return
	}

	if key == "payment-due-days" {
		days, err := h.configService.GetPaymentDueDays()
		if err != nil {
			RespondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		RespondSuccess(w, map[string]int{"days": days})
		return
	}

	if key == "severe-overdue-days" {
		days, err := h.configService.GetSevereOverdueDays()
		if err != nil {
			RespondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		RespondSuccess(w, map[string]int{"days": days})
		return
	}

	config, err := h.configService.GetConfig(key)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if config == nil {
		RespondError(w, http.StatusNotFound, "config not found")
		return
	}

	RespondSuccess(w, config)
}

func (h *ConfigHandler) setConfig(w http.ResponseWriter, r *http.Request) {
	var req SetConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		RespondError(w, http.StatusBadRequest, "key is required")
		return
	}

	if req.Value == "" {
		RespondError(w, http.StatusBadRequest, "value is required")
		return
	}

	if err := h.configService.SetConfig(key, req.Value, req.Description); err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondSuccess(w, map[string]string{"message": "config set successfully"})
}

func (h *ConfigHandler) updateConfig(w http.ResponseWriter, r *http.Request, key string) {
	if key == "sms-unit-price" {
		var req SetPriceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			RespondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Price < 0 {
			RespondError(w, http.StatusBadRequest, "price must be non-negative")
			return
		}

		if err := h.configService.SetSmsUnitPrice(req.Price); err != nil {
			RespondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		RespondSuccess(w, map[string]string{"message": "sms unit price updated successfully"})
		return
	}

	if key == "storage-unit-price" {
		var req SetPriceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			RespondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Price < 0 {
			RespondError(w, http.StatusBadRequest, "price must be non-negative")
			return
		}

		if err := h.configService.SetStorageUnitPrice(req.Price); err != nil {
			RespondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		RespondSuccess(w, map[string]string{"message": "storage unit price updated successfully"})
		return
	}

	var req SetConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Value == "" {
		RespondError(w, http.StatusBadRequest, "value is required")
		return
	}

	if err := h.configService.SetConfig(key, req.Value, req.Description); err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondSuccess(w, map[string]string{"message": "config updated successfully"})
}
