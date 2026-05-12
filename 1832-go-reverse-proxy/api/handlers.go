package api

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"reverse-proxy/store"
	"reverse-proxy/types"
)

type AddBackendRequest struct {
	Hosts   []string `json:"hosts" binding:"required"`
	Backend string   `json:"backend" binding:"required"`
}

type BackendResponse struct {
	ID             string    `json:"id"`
	Hosts          []string  `json:"hosts"`
	Backend        string    `json:"backend"`
	Status         string    `json:"status"`
	LastCheckTime  time.Time `json:"last_check_time"`
	LastChangeTime time.Time `json:"last_change_time"`
}

type Handlers struct {
	store *store.Store
}

func New(store *store.Store) *Handlers {
	return &Handlers{
		store: store,
	}
}

func (h *Handlers) GetBackends(c *gin.Context) {
	backends := h.store.GetAll()
	response := make([]BackendResponse, 0, len(backends))

	for _, b := range backends {
		response = append(response, BackendResponse{
			ID:             b.ID,
			Hosts:          b.Hosts,
			Backend:        b.TargetURL.String(),
			Status:         string(b.GetStatus()),
			LastCheckTime:  b.GetLastCheckTime(),
			LastChangeTime: b.GetLastChangeTime(),
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handlers) AddBackend(c *gin.Context) {
	var req AddBackendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	targetURL, err := url.Parse(req.Backend)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid backend URL"})
		return
	}

	backend := &types.Backend{
		ID:             uuid.New().String(),
		Hosts:          req.Hosts,
		TargetURL:      targetURL,
		Status:         types.StatusHealthy,
		LastCheckTime:  time.Now(),
		LastChangeTime: time.Now(),
	}

	h.store.Add(backend)

	c.JSON(http.StatusCreated, BackendResponse{
		ID:             backend.ID,
		Hosts:          backend.Hosts,
		Backend:        backend.TargetURL.String(),
		Status:         string(backend.Status),
		LastCheckTime:  backend.LastCheckTime,
		LastChangeTime: backend.LastChangeTime,
	})
}

func (h *Handlers) DeleteBackend(c *gin.Context) {
	id := c.Param("id")

	if h.store.GetByID(id) == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "backend not found"})
		return
	}

	h.store.Remove(id)
	c.Status(http.StatusNoContent)
}
