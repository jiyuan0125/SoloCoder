package core

import "time"

type PlaceType string

const (
	PlaceTypeRestaurant   PlaceType = "restaurant"
	PlaceTypeFoodFactory  PlaceType = "food_factory"
	PlaceTypeOffice       PlaceType = "office"
	PlaceTypeHospital     PlaceType = "hospital"
	PlaceTypeSchool       PlaceType = "school"
)

type PestType string

const (
	PestTypeRat     PestType = "rat"
	PestTypeCockroach PestType = "cockroach"
	PestTypeMosquito  PestType = "mosquito"
	PestTypeAnt       PestType = "ant"
	PestTypeTermite   PestType = "termite"
)

type InspectionLevel string

const (
	InspectionLevelNormal  InspectionLevel = "normal"
	InspectionLevelMinor   InspectionLevel = "minor"
	InspectionLevelSevere  InspectionLevel = "severe"
)

type Customer struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Address       string    `json:"address"`
	PlaceType     PlaceType `json:"place_type"`
	Area          float64   `json:"area"`
	PestTypes     []PestType `json:"pest_types"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Contract struct {
	ID                string    `json:"id"`
	CustomerID        string    `json:"customer_id"`
	Customer          *Customer `json:"customer,omitempty"`
	StartDate         time.Time `json:"start_date"`
	EndDate           time.Time `json:"end_date"`
	ServicePerVisit   float64   `json:"service_per_visit"`
	AnnualTotal       float64   `json:"annual_total"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type ControlPoint struct {
	ID          string `json:"id"`
	ContractID  string `json:"contract_id"`
	Code        string `json:"code"`
	Location    string `json:"location"`
	Description string `json:"description"`
}

type PestRecord struct {
	PestType PestType `json:"pest_type"`
	Count    int      `json:"count"`
}

type Inspection struct {
	ID                string           `json:"id"`
	ContractID        string           `json:"contract_id"`
	Contract          *Contract        `json:"contract,omitempty"`
	InspectorID       string           `json:"inspector_id"`
	InspectorName     string           `json:"inspector_name"`
	InspectionDate    time.Time        `json:"inspection_date"`
	ControlPoints     []InspectionPoint `json:"control_points"`
	PestRecords       []PestRecord     `json:"pest_records"`
	Level             InspectionLevel  `json:"level"`
	Notes             string           `json:"notes"`
	RelatedOperationID string          `json:"related_operation_id,omitempty"`
	CreatedAt         time.Time        `json:"created_at"`
}

type InspectionPoint struct {
	ControlPointID   string `json:"control_point_id"`
	ControlPointCode string `json:"control_point_code"`
	FacilityStatus   string `json:"facility_status"`
	PestRecords      []PestRecord `json:"pest_records"`
}

type Chemical struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Unit         string  `json:"unit"`
	UnitPrice    float64 `json:"unit_price"`
	Stock        float64 `json:"stock"`
	SafetyStock  float64 `json:"safety_stock"`
	NeedsPurchase bool   `json:"needs_purchase"`
}

type ChemicalUsage struct {
	ChemicalID string  `json:"chemical_id"`
	Name       string  `json:"name"`
	Amount     float64 `json:"amount"`
	Unit       string  `json:"unit"`
	UnitPrice  float64 `json:"unit_price"`
	TotalCost  float64 `json:"total_cost"`
}

type Staff struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	Phone     string `json:"phone"`
	Active    bool   `json:"active"`
}

type Operation struct {
	ID              string          `json:"id"`
	ContractID      string          `json:"contract_id"`
	Contract        *Contract       `json:"contract,omitempty"`
	OperationDate   time.Time       `json:"operation_date"`
	OperatorID      string          `json:"operator_id"`
	OperatorName    string          `json:"operator_name"`
	Scope           string          `json:"scope"`
	ChemicalUsages  []ChemicalUsage `json:"chemical_usages"`
	TotalChemicalCost float64       `json:"total_chemical_cost"`
	ServiceCost     float64         `json:"service_cost"`
	TotalCost       float64         `json:"total_cost"`
	Notes           string          `json:"notes"`
	Status          string          `json:"status"`
	CreatedAt       time.Time       `json:"created_at"`
}

type EffectEvaluation struct {
	InspectionID       string  `json:"inspection_id"`
	RelatedOperationID string  `json:"related_operation_id"`
	PreviousCount      int     `json:"previous_count"`
	CurrentCount       int     `json:"current_count"`
	ReductionRate      float64 `json:"reduction_rate"`
	IsEffective        bool    `json:"is_effective"`
	NeedsReadjustment  bool    `json:"needs_readjustment"`
}

type PendingService struct {
	ContractID      string    `json:"contract_id"`
	CustomerName    string    `json:"customer_name"`
	PlaceType       PlaceType `json:"place_type"`
	LastServiceDate time.Time `json:"last_service_date"`
	NextServiceDate time.Time `json:"next_service_date"`
	ContractEndDate time.Time `json:"contract_end_date"`
}
