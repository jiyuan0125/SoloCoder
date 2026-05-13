package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	store    *Store
	executor *Executor
}

func NewHandler(store *Store, executor *Executor) *Handler {
	return &Handler{
		store:    store,
		executor: executor,
	}
}

func (h *Handler) CreateBackend(c *gin.Context) {
	var backend Backend
	if err := c.ShouldBindJSON(&backend); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if backend.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "backend id is required"})
		return
	}

	if backend.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "backend url is required"})
		return
	}

	if _, err := url.Parse(backend.URL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid backend url"})
		return
	}

	h.store.AddBackend(&backend)
	c.JSON(http.StatusCreated, backend)
}

func (h *Handler) ListBackends(c *gin.Context) {
	backends := h.store.ListBackends()
	c.JSON(http.StatusOK, backends)
}

func (h *Handler) CreatePlan(c *gin.Context) {
	var plan Plan
	if err := c.ShouldBindJSON(&plan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if plan.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "plan id is required"})
		return
	}

	if len(plan.Services) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "plan must have at least one service"})
		return
	}

	for _, svc := range plan.Services {
		if svc.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "service name is required"})
			return
		}
		if svc.URL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("service %s: url is required", svc.Name)})
			return
		}
		if svc.Mode != ExecutionModeParallel && svc.Mode != ExecutionModeSerial {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("service %s: invalid mode, must be 'parallel' or 'serial'", svc.Name)})
			return
		}
		if svc.FallbackPolicy != FallbackReturnDefault && svc.FallbackPolicy != FallbackIgnore && svc.FallbackPolicy != FallbackFailAll {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("service %s: invalid fallback_policy, must be 'return_default', 'ignore' or 'fail_all'", svc.Name)})
			return
		}
		if svc.FallbackPolicy == FallbackReturnDefault && svc.DefaultValue == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("service %s: default_value is required when fallback_policy is 'return_default'", svc.Name)})
			return
		}
	}

	if plan.Template == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "template is required"})
		return
	}

	h.store.AddPlan(&plan)
	c.JSON(http.StatusCreated, plan)
}

func (h *Handler) ListPlans(c *gin.Context) {
	plans := h.store.ListPlans()
	c.JSON(http.StatusOK, plans)
}

func (h *Handler) GetPlan(c *gin.Context) {
	id := c.Param("id")
	plan, ok := h.store.GetPlan(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
		return
	}
	c.JSON(http.StatusOK, plan)
}

func (h *Handler) ExecutePlan(c *gin.Context) {
	id := c.Param("id")
	plan, ok := h.store.GetPlan(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
		return
	}

	results, err := h.executor.ExecutePlan(plan)
	if err != nil {
		if execErr, ok := err.(*ExecutionError); ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":        "service execution failed",
				"service_name": execErr.ServiceName,
				"details":      execErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	merged := ApplyTemplate(plan.Template, results)
	c.JSON(http.StatusOK, merged)
}
