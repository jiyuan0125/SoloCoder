package models

import (
	"time"

	"gorm.io/gorm"
)

type DosageForm string

const (
	DosageTablet     DosageForm = "tablet"
	DosageCapsule    DosageForm = "capsule"
	DosageInjection  DosageForm = "injection"
	DosageGranule    DosageForm = "granule"
	DosageSyrup      DosageForm = "syrup"
	DosageExternal   DosageForm = "external"
)

type CategoryLevel string

const (
	CategoryChemical     CategoryLevel = "化学药品"
	CategoryChinesePatent CategoryLevel = "中成药"
	CategoryBiological   CategoryLevel = "生物制品"
	CategoryMedical      CategoryLevel = "医疗器械"
)

type DrugCategory struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name"`
	ParentID  *uint          `json:"parent_id"`
	Level     CategoryLevel  `json:"level"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type Drug struct {
	ID                   uint           `json:"id" gorm:"primaryKey"`
	DrugCode             string         `json:"drug_code" gorm:"uniqueIndex;not null"`
	GenericName          string         `json:"generic_name" gorm:"not null"`
	BrandName            string         `json:"brand_name"`
	Specification        string         `json:"specification" gorm:"not null"`
	Manufacturer         string         `json:"manufacturer" gorm:"not null"`
	DosageForm           DosageForm     `json:"dosage_form" gorm:"not null"`
	Unit                 string         `json:"unit" gorm:"not null"`
	RetailPrice          float64        `json:"retail_price" gorm:"not null"`
	PurchasePrice        float64        `json:"purchase_price" gorm:"not null"`
	CategoryID           uint           `json:"category_id"`
	SupportSplit         bool           `json:"support_split" gorm:"default:false"`
	IsSpecialDrug        bool           `json:"is_special_drug" gorm:"default:false"`
	CurrentStock         int            `json:"current_stock" gorm:"default:0"`
	MinStock             int            `json:"min_stock" gorm:"default:0"`
	MaxStock             int            `json:"max_stock" gorm:"default:1000"`
	Category             DrugCategory   `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	StockItems           []StockItem    `json:"stock_items,omitempty" gorm:"foreignKey:DrugID"`
	PrescriptionItems    []PrescriptionItem `json:"-" gorm:"foreignKey:DrugID"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `json:"-" gorm:"index"`
}

type StockItem struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	DrugID         uint           `json:"drug_id" gorm:"not null"`
	BatchNumber    string         `json:"batch_number" gorm:"uniqueIndex;not null"`
	Quantity       int            `json:"quantity" gorm:"not null"`
	ProductionDate time.Time      `json:"production_date" gorm:"not null"`
	ExpiryDate     time.Time      `json:"expiry_date" gorm:"not null"`
	Supplier       string         `json:"supplier" gorm:"not null"`
	SupplierCode   string         `json:"supplier_code" gorm:"not null"`
	Drug           Drug           `json:"drug,omitempty" gorm:"foreignKey:DrugID"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

type StockTransaction struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	TransactionType string         `json:"transaction_type" gorm:"not null"`
	DrugID         uint           `json:"drug_id" gorm:"not null"`
	StockItemID    *uint          `json:"stock_item_id"`
	Quantity       int            `json:"quantity" gorm:"not null"`
	BatchNumber    string         `json:"batch_number"`
	ProductionDate time.Time      `json:"production_date"`
	ExpiryDate     time.Time      `json:"expiry_date"`
	Supplier       string         `json:"supplier"`
	SupplierCode   string         `json:"supplier_code"`
	Operator1      string         `json:"operator1"`
	Operator2      string         `json:"operator2"`
	Remark         string         `json:"remark"`
	PrescriptionID *uint          `json:"prescription_id"`
	Drug           Drug           `json:"drug,omitempty" gorm:"foreignKey:DrugID"`
	CreatedAt      time.Time      `json:"created_at"`
}

type StockAlert struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	DrugID      uint           `json:"drug_id" gorm:"not null"`
	AlertType   string         `json:"alert_type" gorm:"not null"`
	Message     string         `json:"message" gorm:"not null"`
	IsRead      bool           `json:"is_read" gorm:"default:false"`
	Drug        Drug           `json:"drug,omitempty" gorm:"foreignKey:DrugID"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type PrescriptionStatus string

const (
	PrescriptionPending   PrescriptionStatus = "pending"
	PrescriptionApproved  PrescriptionStatus = "approved"
	PrescriptionCancelled PrescriptionStatus = "cancelled"
)

type Prescription struct {
	ID              uint               `json:"id" gorm:"primaryKey"`
	PrescriptionNo  string             `json:"prescription_no" gorm:"uniqueIndex;not null"`
	PatientName     string             `json:"patient_name" gorm:"not null"`
	PatientID       string             `json:"patient_id" gorm:"not null"`
	Diagnosis       string             `json:"diagnosis" gorm:"not null"`
	Status          PrescriptionStatus `json:"status" gorm:"not null;default:pending"`
	ReviewComment   string             `json:"review_comment"`
	Items           []PrescriptionItem `json:"items" gorm:"foreignKey:PrescriptionID"`
	Transactions    []StockTransaction `json:"transactions,omitempty" gorm:"foreignKey:PrescriptionID"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

type PrescriptionItem struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	PrescriptionID uint           `json:"prescription_id" gorm:"not null"`
	DrugID         uint           `json:"drug_id" gorm:"not null"`
	DrugName       string         `json:"drug_name"`
	Quantity       float64        `json:"quantity" gorm:"not null"`
	Usage          string         `json:"usage" gorm:"not null"`
	Drug           Drug           `json:"drug,omitempty" gorm:"foreignKey:DrugID"`
	CreatedAt      time.Time      `json:"created_at"`
}
