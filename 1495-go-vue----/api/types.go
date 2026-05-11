package api

import (
	"pest-control/core"
	"time"
)

type Response struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type CreateCustomerRequest struct {
	Name      string          `json:"name"`
	Address   string          `json:"address"`
	PlaceType core.PlaceType  `json:"place_type"`
	Area      float64         `json:"area"`
	PestTypes []core.PestType `json:"pest_types"`
}

type CreateContractRequest struct {
	CustomerID      string    `json:"customer_id"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	ServicePerVisit float64   `json:"service_per_visit"`
}

type CreateControlPointRequest struct {
	ContractID  string `json:"contract_id"`
	Code        string `json:"code"`
	Location    string `json:"location"`
	Description string `json:"description"`
}

type CreateStaffRequest struct {
	Name  string `json:"name"`
	Role  string `json:"role"`
	Phone string `json:"phone"`
}

type CreateChemicalRequest struct {
	Name        string  `json:"name"`
	Unit        string  `json:"unit"`
	UnitPrice   float64 `json:"unit_price"`
	Stock       float64 `json:"stock"`
	SafetyStock float64 `json:"safety_stock"`
}

type CreateInspectionPoint struct {
	ControlPointID   string            `json:"control_point_id"`
	ControlPointCode string            `json:"control_point_code"`
	FacilityStatus   string            `json:"facility_status"`
	PestRecords      []core.PestRecord `json:"pest_records"`
}

type CreateInspectionRequest struct {
	ContractID     string                   `json:"contract_id"`
	InspectorID    string                   `json:"inspector_id"`
	InspectionDate time.Time                `json:"inspection_date"`
	ControlPoints  []CreateInspectionPoint  `json:"control_points"`
	Notes          string                   `json:"notes"`
}

type ChemicalUsageRequest struct {
	ChemicalID string  `json:"chemical_id"`
	Amount     float64 `json:"amount"`
}

type CreateOperationRequest struct {
	ContractID    string                  `json:"contract_id"`
	OperatorID    string                  `json:"operator_id"`
	OperationDate time.Time               `json:"operation_date"`
	Scope         string                  `json:"scope"`
	Usages        []ChemicalUsageRequest  `json:"usages"`
	Notes         string                  `json:"notes"`
}

type CustomerListResponse struct {
	Customers []*core.Customer `json:"customers"`
}

type ContractListResponse struct {
	Contracts []*core.Contract `json:"contracts"`
}

type StaffListResponse struct {
	Staff []*core.Staff `json:"staff"`
}

type ChemicalListResponse struct {
	Chemicals []*core.Chemical `json:"chemicals"`
}

type InspectionListResponse struct {
	Inspections []*core.Inspection `json:"inspections"`
}

type OperationListResponse struct {
	Operations []*core.Operation `json:"operations"`
}

type ControlPointListResponse struct {
	ControlPoints []*core.ControlPoint `json:"control_points"`
}

type PendingServiceListResponse struct {
	Services []*core.PendingService `json:"services"`
}
