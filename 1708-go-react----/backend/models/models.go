package models

import (
	"blood-management-system/config"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Donor struct {
	ID               string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	DonorNumber      string    `gorm:"uniqueIndex;type:varchar(20)" json:"donor_number"`
	Name             string    `gorm:"type:varchar(100);not null" json:"name"`
	IDCard           string    `gorm:"uniqueIndex;type:varchar(18);not null" json:"id_card"`
	Gender           string    `gorm:"type:varchar(10);not null" json:"gender"`
	BirthDate        time.Time `json:"birth_date"`
	ABOBloodType     string    `gorm:"type:varchar(5);not null" json:"abo_blood_type"`
	RhFactor         string    `gorm:"type:varchar(10);not null" json:"rh_factor"`
	BloodType        string    `gorm:"type:varchar(20);not null" json:"blood_type"`
	Phone            string    `gorm:"type:varchar(20)" json:"phone"`
	Address          string    `gorm:"type:text" json:"address"`
	LastDonationDate *time.Time `json:"last_donation_date"`
	LastDonationType *string    `json:"last_donation_type"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type HealthCheck struct {
	ID                string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	DonorID           string    `gorm:"index;type:varchar(36);not null" json:"donor_id"`
	HeightCm          float64   `json:"height_cm"`
	WeightKg          float64   `json:"weight_kg"`
	RecentMedication  string    `gorm:"type:text" json:"recent_medication"`
	IsFasting         bool      `json:"is_fasting"`
	CheckDate         time.Time `json:"check_date"`
	Notes             string    `gorm:"type:text" json:"notes"`
	CreatedAt         time.Time `json:"created_at"`
}

type BloodCollection struct {
	ID               string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Barcode          string    `gorm:"uniqueIndex;type:varchar(50);not null" json:"barcode"`
	DonorID          string    `gorm:"index;type:varchar(36);not null" json:"donor_id"`
	Donor            *Donor    `gorm:"foreignKey:DonorID" json:"donor,omitempty"`
	CollectionType   string    `gorm:"type:varchar(30);not null" json:"collection_type"`
	VolumeML         float64   `gorm:"not null" json:"volume_ml"`
	ProductType      string    `gorm:"type:varchar(30);not null" json:"product_type"`
	CollectorStaffID string    `gorm:"type:varchar(50)" json:"collector_staff_id"`
	CollectionTime   time.Time `json:"collection_time"`
	ExpiryDate       time.Time `json:"expiry_date"`
	Status           string    `gorm:"type:varchar(30);not null;default:Pending_Test" json:"status"`
	ScrapReason      string    `gorm:"type:text" json:"scrap_reason"`
	TestProgress     int       `gorm:"default:0" json:"test_progress"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type TestRecord struct {
	ID             string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	CollectionID   string    `gorm:"index;type:varchar(36);not null" json:"collection_id"`
	TestRound      int       `gorm:"not null" json:"test_round"`
	ReagentVendor  string    `gorm:"type:varchar(100)" json:"reagent_vendor"`
	ABOFront       string    `gorm:"type:varchar(20)" json:"abo_front"`
	ABOBack        string    `gorm:"type:varchar(20)" json:"abo_back"`
	RhD            string    `gorm:"type:varchar(20)" json:"rhd"`
	ALT            string    `gorm:"type:varchar(20)" json:"alt"`
	HBsAg          string    `gorm:"type:varchar(20)" json:"hbsag"`
	AntiHCV        string    `gorm:"type:varchar(20)" json:"anti_hcv"`
	AntiHIV        string    `gorm:"type:varchar(20)" json:"anti_hiv"`
	AntiSyphilis   string    `gorm:"type:varchar(20)" json:"anti_syphilis"`
	NAT            string    `gorm:"type:varchar(20)" json:"nat"`
	OperatorID     string    `gorm:"type:varchar(50)" json:"operator_id"`
	TestTime       time.Time `json:"test_time"`
	CreatedAt      time.Time `json:"created_at"`
}

type Inventory struct {
	ID             string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	CollectionID   string    `gorm:"uniqueIndex;type:varchar(36);not null" json:"collection_id"`
	Collection     *BloodCollection `gorm:"foreignKey:CollectionID" json:"collection,omitempty"`
	Barcode        string    `gorm:"uniqueIndex;type:varchar(50);not null" json:"barcode"`
	BloodType      string    `gorm:"index;type:varchar(20);not null" json:"blood_type"`
	ProductType    string    `gorm:"index;type:varchar(30);not null" json:"product_type"`
	VolumeML       float64   `json:"volume_ml"`
	StorageDate    time.Time `json:"storage_date"`
	ExpiryDate     time.Time `json:"expiry_date"`
	Status         string    `gorm:"type:varchar(30);not null;default:In_Stock" json:"status"`
	RequestID      *string   `gorm:"index;type:varchar(36)" json:"request_id"`
	IssuedDate     *time.Time `json:"issued_date"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type BloodRequest struct {
	ID            string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	HospitalName  string    `gorm:"type:varchar(200);not null" json:"hospital_name"`
	HospitalID    string    `gorm:"type:varchar(50)" json:"hospital_id"`
	BloodType     string    `gorm:"type:varchar(20);not null" json:"blood_type"`
	ProductType   string    `gorm:"type:varchar(30);not null" json:"product_type"`
	Quantity      int       `gorm:"not null" json:"quantity"`
	Urgency       string    `gorm:"type:varchar(20);not null" json:"urgency"`
	PatientInfo   string    `gorm:"type:text" json:"patient_info"`
	RequestTime   time.Time `json:"request_time"`
	Status        string    `gorm:"type:varchar(30);not null;default:Pending" json:"status"`
	UsedUniversal bool      `gorm:"default:false" json:"used_universal"`
	Notes         string    `gorm:"type:text" json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type WorkflowItem struct {
	ID              string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	RelatedType     string    `gorm:"index;type:varchar(50);not null" json:"related_type"`
	RelatedID       string    `gorm:"index;type:varchar(36);not null" json:"related_id"`
	Status          string    `gorm:"type:varchar(30);not null;default:Draft" json:"status"`
	SubmittedAt     *time.Time `json:"submitted_at"`
	ApprovedAt      *time.Time `json:"approved_at"`
	PublishedAt     *time.Time `json:"published_at"`
	RejectedReason  string    `gorm:"type:text" json:"rejected_reason"`
	ReviewerID      string    `gorm:"type:varchar(50)" json:"reviewer_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type SafetyStock struct {
	ID          string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	BloodType   string    `gorm:"index;type:varchar(20);not null" json:"blood_type"`
	ProductType string    `gorm:"index;type:varchar(30);not null" json:"product_type"`
	MinQuantity int       `gorm:"not null;default:10" json:"min_quantity"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func InitDB() (*gorm.DB, error) {
	dsn := config.GetDBPath()
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func GenerateUUID() string {
	return uuid.New().String()
}

func GenerateDonorNumber(seq int) string {
	return fmt.Sprintf("D%010d", seq)
}

func GenerateBarcode(seq int) string {
	return fmt.Sprintf("BC%012d", seq)
}

func GetExpiryDate(productType string, collectionTime time.Time) time.Time {
	switch productType {
	case config.ProductWholeBlood, config.ProductRBC:
		return collectionTime.AddDate(0, 0, 35)
	case config.ProductFFP:
		return collectionTime.AddDate(1, 0, 0)
	case config.ProductPlatelets:
		return collectionTime.AddDate(0, 0, 5)
	default:
		return collectionTime.AddDate(0, 0, 35)
	}
}

func GetProductType(collectionType string) string {
	switch collectionType {
	case config.CollectionWhole200, config.CollectionWhole400:
		return config.ProductWholeBlood
	case config.CollectionPlatelets:
		return config.ProductPlatelets
	default:
		return config.ProductWholeBlood
	}
}

func GetVolume(collectionType string) float64 {
	switch collectionType {
	case config.CollectionWhole200:
		return 200.0
	case config.CollectionWhole400:
		return 400.0
	case config.CollectionPlatelets:
		return 200.0
	default:
		return 400.0
	}
}

func InitSafetyStock(db *gorm.DB) {
	for _, bt := range config.BloodTypeCombinations {
		for _, pt := range config.ProductTypes {
			var count int64
			db.Model(&SafetyStock{}).Where("blood_type = ? AND product_type = ?", bt, pt).Count(&count)
			if count == 0 {
				db.Create(&SafetyStock{
					ID:          GenerateUUID(),
					BloodType:   bt,
					ProductType: pt,
					MinQuantity: 10,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				})
			}
		}
	}
}
