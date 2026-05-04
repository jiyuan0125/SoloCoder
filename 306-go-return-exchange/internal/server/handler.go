package server

import (
	"encoding/json"
	"net/http"
	"return-exchange/pkg/common"

	"github.com/gorilla/mux"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) SubmitReturn(w http.ResponseWriter, r *http.Request) {
	var req common.SubmitReturnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonResponse(w, http.StatusBadRequest, common.SubmitResponse{
			CommonResponse: common.CommonResponse{
				Success: false,
				Message: "invalid request body: " + err.Error(),
			},
		})
		return
	}

	appID, err := h.service.SubmitReturn(req)
	if err != nil {
		h.jsonResponse(w, http.StatusBadRequest, common.SubmitResponse{
			CommonResponse: common.CommonResponse{
				Success: false,
				Message: err.Error(),
			},
		})
		return
	}

	h.jsonResponse(w, http.StatusOK, common.SubmitResponse{
		CommonResponse: common.CommonResponse{
			Success: true,
		},
		ApplicationID: appID,
	})
}

func (h *Handler) SubmitExchange(w http.ResponseWriter, r *http.Request) {
	var req common.SubmitExchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonResponse(w, http.StatusBadRequest, common.SubmitResponse{
			CommonResponse: common.CommonResponse{
				Success: false,
				Message: "invalid request body: " + err.Error(),
			},
		})
		return
	}

	appID, err := h.service.SubmitExchange(req)
	if err != nil {
		h.jsonResponse(w, http.StatusBadRequest, common.SubmitResponse{
			CommonResponse: common.CommonResponse{
				Success: false,
				Message: err.Error(),
			},
		})
		return
	}

	h.jsonResponse(w, http.StatusOK, common.SubmitResponse{
		CommonResponse: common.CommonResponse{
			Success: true,
		},
		ApplicationID: appID,
	})
}

func (h *Handler) ReviewApplication(w http.ResponseWriter, r *http.Request) {
	var req common.ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonResponse(w, http.StatusBadRequest, common.CommonResponse{
			Success: false,
			Message: "invalid request body: " + err.Error(),
		})
		return
	}

	err := h.service.ReviewApplication(req)
	if err != nil {
		h.jsonResponse(w, http.StatusBadRequest, common.CommonResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	h.jsonResponse(w, http.StatusOK, common.CommonResponse{
		Success: true,
	})
}

func (h *Handler) ProcessRefund(w http.ResponseWriter, r *http.Request) {
	var req common.ProcessRefundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonResponse(w, http.StatusBadRequest, common.ProcessRefundResponse{
			CommonResponse: common.CommonResponse{
				Success: false,
				Message: "invalid request body: " + err.Error(),
			},
		})
		return
	}

	refundID, totalRefund, err := h.service.ProcessRefund(req)
	if err != nil {
		h.jsonResponse(w, http.StatusBadRequest, common.ProcessRefundResponse{
			CommonResponse: common.CommonResponse{
				Success: false,
				Message: err.Error(),
			},
		})
		return
	}

	h.jsonResponse(w, http.StatusOK, common.ProcessRefundResponse{
		CommonResponse: common.CommonResponse{
			Success: true,
		},
		RefundID:    refundID,
		TotalRefund: totalRefund,
	})
}

func (h *Handler) CreateShippingOrder(w http.ResponseWriter, r *http.Request) {
	var req common.CreateShippingOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonResponse(w, http.StatusBadRequest, common.CreateShippingOrderResponse{
			CommonResponse: common.CommonResponse{
				Success: false,
				Message: "invalid request body: " + err.Error(),
			},
		})
		return
	}

	shippingOrderID, err := h.service.CreateShippingOrder(req)
	if err != nil {
		h.jsonResponse(w, http.StatusBadRequest, common.CreateShippingOrderResponse{
			CommonResponse: common.CommonResponse{
				Success: false,
				Message: err.Error(),
			},
		})
		return
	}

	h.jsonResponse(w, http.StatusOK, common.CreateShippingOrderResponse{
		CommonResponse: common.CommonResponse{
			Success: true,
		},
		ShippingOrderID: shippingOrderID,
	})
}

func (h *Handler) CompleteApplication(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["id"]

	err := h.service.CompleteApplication(applicationID)
	if err != nil {
		h.jsonResponse(w, http.StatusBadRequest, common.CommonResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	h.jsonResponse(w, http.StatusOK, common.CommonResponse{
		Success: true,
	})
}

func (h *Handler) GetApplication(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["id"]

	app, err := h.service.GetApplication(applicationID)
	if err != nil {
		h.jsonResponse(w, http.StatusNotFound, common.ApplicationDetailResponse{
			CommonResponse: common.CommonResponse{
				Success: false,
				Message: err.Error(),
			},
		})
		return
	}

	h.jsonResponse(w, http.StatusOK, common.ApplicationDetailResponse{
		CommonResponse: common.CommonResponse{
			Success: true,
		},
		Application: app,
	})
}

func (h *Handler) ListAllApplications(w http.ResponseWriter, r *http.Request) {
	apps := h.service.ListAllApplications()
	h.jsonResponse(w, http.StatusOK, common.ApplicationListResponse{
		CommonResponse: common.CommonResponse{
			Success: true,
		},
		Applications: apps,
	})
}

func (h *Handler) ListUserApplications(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	apps := h.service.ListUserApplications(userID)
	h.jsonResponse(w, http.StatusOK, common.ApplicationListResponse{
		CommonResponse: common.CommonResponse{
			Success: true,
		},
		Applications: apps,
	})
}

func (h *Handler) HandlePriceDifference(w http.ResponseWriter, r *http.Request) {
	var req common.HandlePriceDifferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonResponse(w, http.StatusBadRequest, common.HandlePriceDifferenceResponse{
			CommonResponse: common.CommonResponse{
				Success: false,
				Message: "invalid request body: " + err.Error(),
			},
		})
		return
	}

	action, amount, refundID, err := h.service.HandlePriceDifference(req)
	if err != nil {
		h.jsonResponse(w, http.StatusBadRequest, common.HandlePriceDifferenceResponse{
			CommonResponse: common.CommonResponse{
				Success: false,
				Message: err.Error(),
			},
		})
		return
	}

	app, _ := h.service.GetApplication(req.ApplicationID)
	var priceDiff float64
	if app != nil {
		priceDiff = app.PriceDifference
	}

	h.jsonResponse(w, http.StatusOK, common.HandlePriceDifferenceResponse{
		CommonResponse: common.CommonResponse{
			Success: true,
		},
		Action:          action,
		Amount:          amount,
		RefundID:        refundID,
		PriceDifference: priceDiff,
	})
}

func (h *Handler) CreateMockOrder(w http.ResponseWriter, r *http.Request) {
	orderID := h.service.CreateMockOrder()
	h.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"order_id": orderID,
	})
}
