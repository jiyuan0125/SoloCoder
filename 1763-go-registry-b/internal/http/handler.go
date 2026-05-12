package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"registry/internal/registry"
)

type Handler struct {
	reg *registry.Registry
}

func NewHandler(reg *registry.Registry) *Handler {
	return &Handler{reg: reg}
}

type RegisterRequest struct {
	ID          string            `json:"id" binding:"required"`
	ServiceName string            `json:"service_name" binding:"required"`
	IP          string            `json:"ip" binding:"required"`
	Port        int               `json:"port" binding:"required,min=1,max=65535"`
	Metadata    map[string]string `json:"metadata"`
}

type HeartbeatRequest struct {
	ID          string `json:"id" binding:"required"`
	ServiceName string `json:"service_name" binding:"required"`
}

type UnregisterRequest struct {
	ID          string `json:"id" binding:"required"`
	ServiceName string `json:"service_name" binding:"required"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	instance := &registry.ServiceInstance{
		ID:          req.ID,
		ServiceName: req.ServiceName,
		IP:          req.IP,
		Port:        req.Port,
		Metadata:    req.Metadata,
	}

	if err := h.reg.Register(instance); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "registered"})
}

func (h *Handler) Heartbeat(c *gin.Context) {
	var req HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.reg.Heartbeat(req.ServiceName, req.ID); err != nil {
		if errors.Is(err, registry.ErrServiceNotFound) || errors.Is(err, registry.ErrInstanceNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) Unregister(c *gin.Context) {
	var req UnregisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.reg.Unregister(req.ServiceName, req.ID); err != nil {
		if errors.Is(err, registry.ErrServiceNotFound) || errors.Is(err, registry.ErrInstanceNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "unregistered"})
}

func (h *Handler) GetService(c *gin.Context) {
	serviceName := c.Param("service_name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "service_name is required"})
		return
	}

	instances, err := h.reg.GetService(serviceName)
	if err != nil {
		if errors.Is(err, registry.ErrServiceNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	result := make([]gin.H, 0, len(instances))
	for _, inst := range instances {
		result = append(result, gin.H{
			"id":             inst.ID,
			"ip":             inst.IP,
			"port":           inst.Port,
			"metadata":       inst.Metadata,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"service_name": serviceName,
		"instances":    result,
		"count":        len(result),
	})
}

func (h *Handler) GetAllServices(c *gin.Context) {
	services := h.reg.GetAllServices()
	c.JSON(http.StatusOK, gin.H{
		"services": services,
		"count":    len(services),
	})
}

func (h *Handler) Subscribe(c *gin.Context) {
	serviceName := c.Query("service")

	ch := h.reg.Subscribe()
	defer h.reg.Unsubscribe(ch)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	closeNotify := c.Writer.CloseNotify()
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "streaming unsupported"})
		return
	}

	c.Writer.WriteHeader(http.StatusOK)
	flusher.Flush()

	for {
		select {
		case <-closeNotify:
			return
		case change, ok := <-ch:
			if !ok {
				return
			}
			if serviceName != "" && change.ServiceName != serviceName {
				continue
			}
			data, err := json.Marshal(change)
			if err != nil {
				continue
			}
			_, _ = io.WriteString(c.Writer, "event: change\n")
			_, _ = io.WriteString(c.Writer, "data: "+string(data)+"\n\n")
			flusher.Flush()
		}
	}
}

var (
	jsonPool = sync.Pool{
		New: func() any {
			return struct{}{}
		},
	}
)
