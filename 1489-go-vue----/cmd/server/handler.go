package main

import (
	"encoding/json"
	"net/http"

	"lockservice/pkg/common"
	"lockservice/pkg/lockservice"
)

type Handler struct {
	service *lockservice.Service
}

func NewHandler(service *lockservice.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, common.ErrorResp{Success: false, Message: message})
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req common.CreateOrderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	date, err := common.ParseDate(req.TimeSlotDate)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid date format, expected YYYY-MM-DD")
		return
	}

	coreReq := &lockservice.CreateOrderRequest{
		UserID:       req.UserID,
		Address:      req.Address,
		LockType:     lockservice.LockType(req.LockType),
		Urgency:      lockservice.Urgency(req.Urgency),
		TimeSlot:     lockservice.TimeSlot(req.TimeSlot),
		TimeSlotDate: date,
	}

	result := h.service.CreateOrder(coreReq)

	resp := common.CreateOrderResp{
		Success: result.Success,
		OrderID: result.OrderID,
		Message: result.Message,
	}

	if result.Order != nil {
		resp.Order = toOrderView(result.Order)
	}

	if result.Alternative != nil {
		alt := &common.AlternativeView{
			OtherSlots:   make([]string, len(result.Alternative.OtherSlots)),
			OtherMasters: make([]common.MasterSummary, len(result.Alternative.OtherMasters)),
		}
		for i, s := range result.Alternative.OtherSlots {
			alt.OtherSlots[i] = string(s)
		}
		for i, m := range result.Alternative.OtherMasters {
			slots := h.service.MasterManager.FindAvailableSlots(m.ID, date)
			slotStrs := make([]string, len(slots))
			for j, s := range slots {
				slotStrs[j] = string(s)
			}
			alt.OtherMasters[i] = common.MasterSummary{
				ID:            m.ID,
				Name:          m.Name,
				ServiceArea:   m.ServiceArea,
				Status:        string(m.Status),
				AvailableSlots: slotStrs,
			}
		}
		resp.Alternative = alt
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) AcceptOrder(w http.ResponseWriter, r *http.Request) {
	var req common.AcceptOrderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.service.AcceptOrder(req.OrderID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, common.AcceptOrderResp{Success: true})
}

func (h *Handler) StartService(w http.ResponseWriter, r *http.Request) {
	var req common.StartServiceReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.service.StartService(req.OrderID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, common.StartServiceResp{Success: true})
}

func (h *Handler) CompleteService(w http.ResponseWriter, r *http.Request) {
	var req common.CompleteServiceReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	coreReq := &lockservice.CompleteServiceRequest{
		OrderID:         req.OrderID,
		NeedReplaceLock: req.NeedReplaceLock,
		LockBrand:       req.LockBrand,
		LockModel:       req.LockModel,
		LockLevel:       req.LockLevel,
		PartsCost:       req.PartsCost,
	}

	order, err := h.service.CompleteService(coreReq)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, common.CompleteServiceResp{
		Success: true,
		Order:   toOrderView(order),
	})
}

func (h *Handler) PayOrder(w http.ResponseWriter, r *http.Request) {
	var req common.PayOrderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.service.PayOrder(req.OrderID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, common.PayOrderResp{Success: true})
}

func (h *Handler) RateOrder(w http.ResponseWriter, r *http.Request) {
	var req common.RateOrderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result := h.service.RateOrder(&lockservice.RateOrderRequest{
		OrderID: req.OrderID,
		Rating:  req.Rating,
		Comment: req.Comment,
	})

	h.writeJSON(w, http.StatusOK, common.RateOrderResp{
		Success:      result.Success,
		HasComplaint: result.HasComplaint,
		Message:      result.Message,
	})
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "missing order id")
		return
	}

	order := h.service.GetOrder(id)
	if order == nil {
		h.writeError(w, http.StatusNotFound, "order not found")
		return
	}

	h.writeJSON(w, http.StatusOK, common.GetOrderResp{
		Success: true,
		Order:   toOrderView(order),
	})
}

func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	orders := h.service.ListOrders()
	views := make([]*common.OrderView, len(orders))
	for i, o := range orders {
		views[i] = toOrderView(o)
	}

	h.writeJSON(w, http.StatusOK, common.ListOrdersResp{
		Success: true,
		Orders:  views,
	})
}

func (h *Handler) ListMasters(w http.ResponseWriter, r *http.Request) {
	masters := h.service.ListMasters()
	views := make([]*common.MasterView, len(masters))
	for i, m := range masters {
		locks := make([]string, len(m.SupportedLocks))
		for j, l := range m.SupportedLocks {
			locks[j] = string(l)
		}
		views[i] = &common.MasterView{
			ID:             m.ID,
			Name:           m.Name,
			Phone:          m.Phone,
			ServiceArea:    m.ServiceArea,
			SupportedLocks: locks,
			Status:         string(m.Status),
			Location:       m.Location,
		}
	}

	h.writeJSON(w, http.StatusOK, common.ListMastersResp{
		Success: true,
		Masters: views,
	})
}

func (h *Handler) AddMaster(w http.ResponseWriter, r *http.Request) {
	var req common.AddMasterReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	locks := make([]lockservice.LockType, len(req.SupportedLocks))
	for i, l := range req.SupportedLocks {
		locks[i] = lockservice.LockType(l)
	}

	master := &lockservice.Master{
		ID:             req.ID,
		Name:           req.Name,
		Phone:          req.Phone,
		ServiceArea:    req.ServiceArea,
		SupportedLocks: locks,
		Status:         lockservice.MasterStatusIdle,
		Location:       req.Location,
		BookedSlots:    make(map[string]map[lockservice.TimeSlot]bool),
	}

	h.service.AddMaster(master)

	h.writeJSON(w, http.StatusOK, common.AddMasterResp{Success: true})
}

func toOrderView(o *lockservice.Order) *common.OrderView {
	if o == nil {
		return nil
	}
	return &common.OrderView{
		ID:           o.ID,
		UserID:       o.UserID,
		Address:      o.Address,
		LockType:     string(o.LockType),
		Urgency:      string(o.Urgency),
		TimeSlot:     string(o.TimeSlot),
		TimeSlotDate: common.FormatDate(o.TimeSlotDate),
		MasterID:     o.MasterID,
		Status:       string(o.Status),
		Detail: common.DetailView{
			NeedReplaceLock: o.Detail.NeedReplaceLock,
			LockBrand:       o.Detail.LockBrand,
			LockModel:       o.Detail.LockModel,
			LockLevel:       o.Detail.LockLevel,
			PartsCost:       o.Detail.PartsCost,
			LaborCost:       o.Detail.LaborCost,
		},
		BaseCost:     o.BaseCost,
		UrgentFee:    o.UrgentFee,
		TotalCost:    o.TotalCost,
		CreatedAt:    common.FormatDateTime(o.CreatedAt),
		Rating:       o.Rating,
		Comment:      o.Comment,
		HasComplaint: o.HasComplaint,
	}
}
