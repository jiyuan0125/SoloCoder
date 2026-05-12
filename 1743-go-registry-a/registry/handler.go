package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type Handler struct {
	reg *Registry
}

func NewHandler(reg *Registry) *Handler {
	return &Handler{reg: reg}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var inst ServiceInstance
	if err := json.NewDecoder(r.Body).Decode(&inst); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.reg.Register(inst); err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req struct {
		Name string `json:"name"`
		IP   string `json:"ip"`
		Port int    `json:"port"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.reg.Heartbeat(req.Name, req.IP, req.Port); err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) Discover(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	name := r.URL.Query().Get("name")
	version := r.URL.Query().Get("version")

	instances, err := h.reg.Discover(name, version)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, instances)
}

func (h *Handler) Watch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "streaming not supported"})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch, unsubscribe := h.reg.Watch(name)
	defer unsubscribe()

	ctx := r.Context()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: %s\n", event.Action)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	h.reg.mu.RLock()
	defer h.reg.mu.RUnlock()

	result := make(map[string][]ServiceInstance)
	for name, instances := range h.reg.services {
		list := make([]ServiceInstance, 0, len(instances))
		for _, inst := range instances {
			list = append(list, *inst)
		}
		result[name] = list
	}

	WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	name := r.URL.Query().Get("name")
	ip := r.URL.Query().Get("ip")
	portStr := r.URL.Query().Get("port")

	if name == "" || ip == "" || portStr == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "name, ip, port are required"})
		return
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid port"})
		return
	}

	key := ip + ":" + strconv.Itoa(port)

	h.reg.mu.RLock()
	defer h.reg.mu.RUnlock()

	instances, ok := h.reg.services[name]
	if !ok {
		WriteJSON(w, http.StatusNotFound, map[string]string{"error": "service not found"})
		return
	}

	inst, ok := instances[key]
	if !ok {
		WriteJSON(w, http.StatusNotFound, map[string]string{"error": "instance not found"})
		return
	}

	WriteJSON(w, http.StatusOK, inst)
}
