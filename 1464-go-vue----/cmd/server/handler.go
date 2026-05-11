package main

import (
	"encoding/json"
	"errors"
	"firemanagement/internal/core"
	"firemanagement/pkg/api"
	"io"
	"net/http"
	"strings"
)

type Handler struct {
	service *core.Service
}

func NewHandler(service *core.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(api.Response{
			Success: status >= 200 && status < 300,
			Data:    data,
		})
	}
}

func (h *Handler) writeError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(api.Response{
		Success: false,
		Error: &api.ErrorInfo{
			Code:    strings.ToLower(http.StatusText(status)),
			Message: err.Error(),
		},
	})
}

func (h *Handler) readJSON(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if len(body) == 0 {
		return nil
	}
	return json.Unmarshal(body, v)
}

func getIDFromPath(r *http.Request, prefix string) (string, bool) {
	path := r.URL.Path
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	id := strings.TrimPrefix(path, prefix)
	if id == "" {
		return "", false
	}
	return id, true
}

func (h *Handler) HandleCreateDevice(w http.ResponseWriter, r *http.Request) {
	var req api.CreateDeviceRequest
	if err := h.readJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, err)
		return
	}

	id, err := h.service.DeviceService.CreateDevice(req)
	if err != nil {
		if errors.Is(err, core.ErrDuplicateDeviceCode) || errors.Is(err, core.ErrDuplicateActiveDevice) {
			h.writeError(w, http.StatusConflict, err)
			return
		}
		if errors.Is(err, core.ErrInvalidDeviceType) {
			h.writeError(w, http.StatusBadRequest, err)
			return
		}
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	h.writeJSON(w, http.StatusCreated, api.IDResponse{ID: id})
}

func (h *Handler) HandleGetDevice(w http.ResponseWriter, r *http.Request) {
	id, ok := getIDFromPath(r, "/api/devices/")
	if !ok {
		h.writeError(w, http.StatusBadRequest, errors.New("invalid device id"))
		return
	}

	device, err := h.service.DeviceService.GetDevice(id)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			h.writeError(w, http.StatusNotFound, err)
			return
		}
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	h.writeJSON(w, http.StatusOK, api.DeviceResponse{Device: device})
}

func (h *Handler) HandleListDevices(w http.ResponseWriter, r *http.Request) {
	var req api.ListDevicesRequest
	h.readJSON(r, &req)

	devices := h.service.DeviceService.ListDevices(req)
	h.writeJSON(w, http.StatusOK, api.DevicesResponse{Devices: devices})
}

func (h *Handler) HandleUpdateDevice(w http.ResponseWriter, r *http.Request) {
	id, ok := getIDFromPath(r, "/api/devices/")
	if !ok {
		h.writeError(w, http.StatusBadRequest, errors.New("invalid device id"))
		return
	}

	var req api.UpdateDeviceRequest
	if err := h.readJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, err)
		return
	}

	err := h.service.DeviceService.UpdateDevice(id, req)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			h.writeError(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, core.ErrDuplicateDeviceCode) || errors.Is(err, core.ErrDuplicateActiveDevice) {
			h.writeError(w, http.StatusConflict, err)
			return
		}
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	h.writeJSON(w, http.StatusOK, nil)
}

func (h *Handler) HandleCreatePoint(w http.ResponseWriter, r *http.Request) {
	var req api.CreateInspectionPointRequest
	if err := h.readJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, err)
		return
	}

	id := h.service.InspectionService.CreateInspectionPoint(req)
	h.writeJSON(w, http.StatusCreated, api.IDResponse{ID: id})
}

func (h *Handler) HandleListPoints(w http.ResponseWriter, r *http.Request) {
	points := h.service.InspectionService.ListInspectionPoints()
	h.writeJSON(w, http.StatusOK, api.InspectionPointsResponse{Points: points})
}

func (h *Handler) HandleCreateRoute(w http.ResponseWriter, r *http.Request) {
	var req api.CreateInspectionRouteRequest
	if err := h.readJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, err)
		return
	}

	id, err := h.service.InspectionService.CreateInspectionRoute(req)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			h.writeError(w, http.StatusNotFound, err)
			return
		}
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	h.writeJSON(w, http.StatusCreated, api.IDResponse{ID: id})
}

func (h *Handler) HandleListRoutes(w http.ResponseWriter, r *http.Request) {
	routes := h.service.InspectionService.ListInspectionRoutes()
	h.writeJSON(w, http.StatusOK, api.InspectionRoutesResponse{Routes: routes})
}

func (h *Handler) HandleCreatePlan(w http.ResponseWriter, r *http.Request) {
	var req api.CreateInspectionPlanRequest
	if err := h.readJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, err)
		return
	}

	id, err := h.service.InspectionService.CreateInspectionPlan(req)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			h.writeError(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, core.ErrInvalidFrequency) {
			h.writeError(w, http.StatusBadRequest, err)
			return
		}
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	h.writeJSON(w, http.StatusCreated, api.IDResponse{ID: id})
}

func (h *Handler) HandleListPlans(w http.ResponseWriter, r *http.Request) {
	plans := h.service.InspectionService.ListInspectionPlans()
	h.writeJSON(w, http.StatusOK, api.InspectionPlansResponse{Plans: plans})
}

func (h *Handler) HandleListTasks(w http.ResponseWriter, r *http.Request) {
	tasks := h.service.InspectionService.ListInspectionTasks()
	h.writeJSON(w, http.StatusOK, api.InspectionTasksResponse{Tasks: tasks})
}

func (h *Handler) HandleGetTask(w http.ResponseWriter, r *http.Request) {
	id, ok := getIDFromPath(r, "/api/inspection/tasks/")
	if !ok {
		h.writeError(w, http.StatusBadRequest, errors.New("invalid task id"))
		return
	}

	task, err := h.service.InspectionService.GetInspectionTask(id)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			h.writeError(w, http.StatusNotFound, err)
			return
		}
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	h.writeJSON(w, http.StatusOK, api.InspectionTaskResponse{Task: task})
}

func (h *Handler) HandleCheckTaskPoint(w http.ResponseWriter, r *http.Request) {
	var req api.CheckTaskPointRequest
	if err := h.readJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, err)
		return
	}

	err := h.service.InspectionService.CheckTaskPoint(req.TaskID, req.PointID, req.Normal)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			h.writeError(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, core.ErrInvalidPointOrder) || errors.Is(err, core.ErrPointAlreadyChecked) ||
			errors.Is(err, core.ErrTaskAlreadyCompleted) {
			h.writeError(w, http.StatusBadRequest, err)
			return
		}
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	h.writeJSON(w, http.StatusOK, nil)
}

func (h *Handler) HandleTaskSummary(w http.ResponseWriter, r *http.Request) {
	id, ok := getIDFromPath(r, "/api/inspection/tasks/")
	if !ok {
		h.writeError(w, http.StatusBadRequest, errors.New("invalid task id"))
		return
	}
	id = strings.TrimSuffix(id, "/summary")

	summary, err := h.service.InspectionService.GetTaskSummary(id)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			h.writeError(w, http.StatusNotFound, err)
			return
		}
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	h.writeJSON(w, http.StatusOK, summary)
}

func (h *Handler) HandleCreateDrillPlan(w http.ResponseWriter, r *http.Request) {
	var req api.CreateDrillPlanRequest
	if err := h.readJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, err)
		return
	}

	id, err := h.service.DrillService.CreateDrillPlan(req)
	if err != nil {
		if errors.Is(err, core.ErrInvalidDrillType) {
			h.writeError(w, http.StatusBadRequest, err)
			return
		}
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	h.writeJSON(w, http.StatusCreated, api.IDResponse{ID: id})
}

func (h *Handler) HandleListDrillPlans(w http.ResponseWriter, r *http.Request) {
	plans := h.service.DrillService.ListDrillPlans()
	h.writeJSON(w, http.StatusOK, api.DrillPlansResponse{Plans: plans})
}

func (h *Handler) HandleCompleteDrill(w http.ResponseWriter, r *http.Request) {
	id, ok := getIDFromPath(r, "/api/drills/plans/")
	if !ok {
		h.writeError(w, http.StatusBadRequest, errors.New("invalid drill plan id"))
		return
	}
	id = strings.TrimSuffix(id, "/complete")

	var req api.CompleteDrillRequest
	if err := h.readJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, err)
		return
	}

	recordID, err := h.service.DrillService.CompleteDrill(id, req)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			h.writeError(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, core.ErrPlanAlreadyCompleted) {
			h.writeError(w, http.StatusConflict, err)
			return
		}
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	h.writeJSON(w, http.StatusCreated, api.IDResponse{ID: recordID})
}

func (h *Handler) HandleListDrillRecords(w http.ResponseWriter, r *http.Request) {
	records := h.service.DrillService.ListDrillRecords()
	h.writeJSON(w, http.StatusOK, api.DrillRecordsResponse{Records: records})
}

func (h *Handler) HandleListReminders(w http.ResponseWriter, r *http.Request) {
	var req api.ListRemindersRequest
	h.readJSON(r, &req)

	reminders := h.service.DrillService.ListReminders(req)
	h.writeJSON(w, http.StatusOK, api.RemindersResponse{Reminders: reminders})
}

func (h *Handler) HandleMarkReminderRead(w http.ResponseWriter, r *http.Request) {
	id, ok := getIDFromPath(r, "/api/reminders/")
	if !ok {
		h.writeError(w, http.StatusBadRequest, errors.New("invalid reminder id"))
		return
	}
	id = strings.TrimSuffix(id, "/read")

	var req api.MarkReminderReadRequest
	if err := h.readJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, err)
		return
	}

	err := h.service.DrillService.MarkReminderRead(id, req.Read)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			h.writeError(w, http.StatusNotFound, err)
			return
		}
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	h.writeJSON(w, http.StatusOK, nil)
}
