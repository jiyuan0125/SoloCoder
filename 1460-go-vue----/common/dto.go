package common

import (
	"supplier-portal/core"
)

type MaterialDTO struct {
	ID       string  `json:"id,omitempty"`
	Name     string  `json:"name"`
	Spec     string  `json:"spec"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
}

type CreateInquiryRequest struct {
	Title       string        `json:"title"`
	BuyerID     string        `json:"buyer_id"`
	Materials   []MaterialDTO `json:"materials"`
	SupplierIDs []string      `json:"supplier_ids"`
	Deadline    string        `json:"deadline"`
}

type InquiryResponse struct {
	ID               string         `json:"id"`
	Title            string         `json:"title"`
	BuyerID          string         `json:"buyer_id"`
	Materials        []MaterialDTO  `json:"materials"`
	SupplierIDs      []string       `json:"supplier_ids"`
	Deadline         string         `json:"deadline"`
	CreatedAt        string         `json:"created_at"`
	UpdatedAt        string         `json:"updated_at"`
	Status           string         `json:"status"`
	MaterialsVersion int64          `json:"materials_version"`
}

type QuotationItemDTO struct {
	MaterialID   string  `json:"material_id"`
	UnitPrice    float64 `json:"unit_price"`
	DeliveryDays int     `json:"delivery_days"`
}

type SubmitQuotationRequest struct {
	InquiryID  string            `json:"inquiry_id"`
	SupplierID string            `json:"supplier_id"`
	Items      []QuotationItemDTO `json:"items"`
	Remark     string            `json:"remark"`
}

type QuotationResponse struct {
	ID               string             `json:"id"`
	InquiryID        string             `json:"inquiry_id"`
	SupplierID       string             `json:"supplier_id"`
	Items            []QuotationItemDTO `json:"items"`
	TotalPrice       float64            `json:"total_price"`
	Remark           string             `json:"remark"`
	SubmittedAt      string             `json:"submitted_at"`
	Status           string             `json:"status"`
	MaterialsVersion int64              `json:"materials_version"`
}

type ComparisonRequest struct {
	InquiryID string `json:"inquiry_id"`
}

type ComparisonItemResponse struct {
	MaterialID     string             `json:"material_id"`
	MaterialName   string             `json:"material_name"`
	Spec           string             `json:"spec"`
	Quantity       float64            `json:"quantity"`
	Unit           string             `json:"unit"`
	SupplierPrices map[string]float64 `json:"supplier_prices"`
	HasNoPrice     bool               `json:"has_no_price"`
}

type ComparisonResponse struct {
	InquiryID string                  `json:"inquiry_id"`
	Items     []ComparisonItemResponse `json:"items"`
	GeneratedAt string                `json:"generated_at"`
}

type AwardedMaterialDTO struct {
	MaterialID string `json:"material_id"`
	SupplierID string `json:"supplier_id"`
}

type AwardRequest struct {
	InquiryID string               `json:"inquiry_id"`
	Awards    []AwardedMaterialDTO `json:"awards"`
}

type PurchaseOrderItemDTO struct {
	MaterialID   string  `json:"material_id"`
	Name         string  `json:"name"`
	Spec         string  `json:"spec"`
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit"`
	UnitPrice    float64 `json:"unit_price"`
	Amount       float64 `json:"amount"`
	DeliveryDays int     `json:"delivery_days"`
}

type PurchaseOrderResponse struct {
	ID          string                  `json:"id"`
	InquiryID   string                  `json:"inquiry_id"`
	SupplierID  string                  `json:"supplier_id"`
	BuyerID     string                  `json:"buyer_id"`
	Items       []PurchaseOrderItemDTO  `json:"items"`
	TotalAmount float64                 `json:"total_amount"`
	CreatedAt   string                  `json:"created_at"`
	Remark      string                  `json:"remark"`
}

type TodoReminderResponse struct {
	ID           string `json:"id"`
	SupplierID   string `json:"supplier_id"`
	InquiryID    string `json:"inquiry_id"`
	InquiryTitle string `json:"inquiry_title"`
	Deadline     string `json:"deadline"`
	CreatedAt    string `json:"created_at"`
}

type UpdateInquiryMaterialsRequest struct {
	InquiryID string        `json:"inquiry_id"`
	Materials []MaterialDTO `json:"materials"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func ToMaterialDTO(m core.Material) MaterialDTO {
	return MaterialDTO{
		ID:       m.ID,
		Name:     m.Name,
		Spec:     m.Spec,
		Quantity: m.Quantity,
		Unit:     m.Unit,
	}
}

func ToMaterialDTOs(ms []core.Material) []MaterialDTO {
	result := make([]MaterialDTO, len(ms))
	for i, m := range ms {
		result[i] = ToMaterialDTO(m)
	}
	return result
}

func FromMaterialDTO(dto MaterialDTO) core.Material {
	return core.Material{
		ID:       dto.ID,
		Name:     dto.Name,
		Spec:     dto.Spec,
		Quantity: dto.Quantity,
		Unit:     dto.Unit,
	}
}

func FromMaterialDTOs(dtos []MaterialDTO) []core.Material {
	result := make([]core.Material, len(dtos))
	for i, dto := range dtos {
		result[i] = FromMaterialDTO(dto)
	}
	return result
}
