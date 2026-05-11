package main

import (
	"smart-park/core/access"
	"smart-park/core/patrol"
	"smart-park/core/visitor"
)

type Handler struct {
	accessService  *access.Service
	visitorService *visitor.Service
	patrolService  *patrol.Service
}

func NewHandler(
	accessSvc *access.Service,
	visitorSvc *visitor.Service,
	patrolSvc *patrol.Service,
) *Handler {
	return &Handler{
		accessService:  accessSvc,
		visitorService: visitorSvc,
		patrolService:  patrolSvc,
	}
}
