package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"not null"`
	Name      string         `json:"name" gorm:"not null"`
	Role      string         `json:"role" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type Patient struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	Name           string         `json:"name" gorm:"not null"`
	Gender         string         `json:"gender" gorm:"not null"`
	Age            int            `json:"age" gorm:"not null"`
	IDCard         string         `json:"id_card" gorm:"uniqueIndex;not null"`
	Phone          string         `json:"phone" gorm:"not null"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
	Consultations  []Consultation `json:"-"`
}

type Consultation struct {
	ID                   uint           `json:"id" gorm:"primaryKey"`
	ConsultationNo       string         `json:"consultation_no" gorm:"uniqueIndex;not null"`
	PatientID            uint           `json:"patient_id" gorm:"not null"`
	Patient              Patient        `json:"patient" gorm:"foreignKey:PatientID"`
	GrassrootDoctorID    uint           `json:"grassroot_doctor_id" gorm:"not null"`
	GrassrootDoctor      User           `json:"grassroot_doctor" gorm:"foreignKey:GrassrootDoctorID"`
	ExpertDoctorID       *uint          `json:"expert_doctor_id"`
	ExpertDoctor         *User          `json:"expert_doctor" gorm:"foreignKey:ExpertDoctorID"`
	ChiefComplaint       string         `json:"chief_complaint" gorm:"not null"`
	PastHistory          string         `json:"past_history"`
	Status               string         `json:"status" gorm:"not null;default:pending"`
	Diagnosis            string         `json:"diagnosis"`
	PrescriptionAdvice   string         `json:"prescription_advice"`
	TreatmentAdvice      string         `json:"treatment_advice"`
	AcceptedAt           *time.Time     `json:"accepted_at"`
	CompletedAt          *time.Time     `json:"completed_at"`
	Version              int            `json:"version" gorm:"not null;default:0"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `json:"-" gorm:"index"`
	Exams                []Exam         `json:"-"`
	Messages             []Message      `json:"-"`
	Prescriptions        []Prescription `json:"-"`
	ImageDatas           []ImageData    `json:"-"`
}

type Exam struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	ConsultationID uint           `json:"consultation_id" gorm:"not null"`
	ExamType       string         `json:"exam_type" gorm:"not null"`
	ExamResult     string         `json:"exam_result" gorm:"not null"`
	ExamDate       *time.Time     `json:"exam_date"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

type Message struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	ConsultationID uint           `json:"consultation_id" gorm:"not null"`
	SenderID       uint           `json:"sender_id" gorm:"not null"`
	SenderRole     string         `json:"sender_role" gorm:"not null"`
	SenderName     string         `json:"sender_name" gorm:"not null"`
	Content        string         `json:"content" gorm:"not null"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

type Prescription struct {
	ID             uint               `json:"id" gorm:"primaryKey"`
	ConsultationID uint               `json:"consultation_id" gorm:"not null"`
	Consultation   Consultation       `json:"-" gorm:"foreignKey:ConsultationID"`
	PrescriptionNo string             `json:"prescription_no" gorm:"uniqueIndex;not null"`
	TotalAmount    float64            `json:"total_amount"`
	Status         string             `json:"status" gorm:"not null;default:active"`
	ExpiredAt      time.Time          `json:"expired_at"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
	DeletedAt      gorm.DeletedAt     `json:"-" gorm:"index"`
	Items          []PrescriptionItem `json:"items" gorm:"foreignKey:PrescriptionID"`
}

type PrescriptionItem struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	PrescriptionID uint           `json:"prescription_id" gorm:"not null"`
	DrugName       string         `json:"drug_name" gorm:"not null"`
	Specification  string         `json:"specification" gorm:"not null"`
	Usage          string         `json:"usage" gorm:"not null"`
	Dosage         string         `json:"dosage" gorm:"not null"`
	Days           int            `json:"days" gorm:"not null"`
	UnitPrice      float64        `json:"unit_price"`
	Quantity       int            `json:"quantity"`
	Amount         float64        `json:"amount"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

type ForbiddenDrug struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"uniqueIndex;not null"`
	Category  string         `json:"category" gorm:"not null"`
	Reason    string         `json:"reason"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type ImageTemplate struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Type      string         `json:"type" gorm:"uniqueIndex;not null"`
	Fields    string         `json:"fields" gorm:"not null"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type ImageData struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	ConsultationID uint           `json:"consultation_id" gorm:"not null"`
	TemplateID     uint           `json:"template_id" gorm:"not null"`
	Template       ImageTemplate  `json:"template" gorm:"foreignKey:TemplateID"`
	ImageType      string         `json:"image_type" gorm:"not null"`
	FieldValues    string         `json:"field_values" gorm:"not null"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

type DrugPriceHistory struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	DrugName     string         `json:"drug_name" gorm:"not null"`
	OldPrice     float64        `json:"old_price"`
	NewPrice     float64        `json:"new_price"`
	ChangeReason string         `json:"change_reason"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}
