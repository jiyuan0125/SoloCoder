package models

import "time"

type PatientStatus string

const (
	StatusWaiting PatientStatus = "waiting"
	StatusInVisit PatientStatus = "in_visit"
	StatusDone    PatientStatus = "done"
)

type Registration struct {
	ID        string        `json:"id"`
	PatientID string        `json:"patient_id"`
	SerialNum string        `json:"serial_num"`
	Date      string        `json:"date"`
	Status    PatientStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
}

type Patient struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Phone        string    `json:"phone"`
	ChiefComplaint string  `json:"chief_complaint"`
	CreatedAt    time.Time `json:"created_at"`
}

type PrescriptionStatus string

const (
	PrescriptionDraft  PrescriptionStatus = "draft"
	PrescriptionConfirmed PrescriptionStatus = "confirmed"
	PrescriptionDispensed PrescriptionStatus = "dispensed"
)

type PrescriptionItem struct {
	HerbName     string  `json:"herb_name"`
	Dosage       float64 `json:"dosage"`
	Unit         string  `json:"unit"`
	CookingMethod string `json:"cooking_method"`
}

type Prescription struct {
	ID                 string             `json:"id"`
	RegistrationID     string             `json:"registration_id"`
	PatientID          string             `json:"patient_id"`
	PatientName        string             `json:"patient_name"`
	Diagnosis          string             `json:"diagnosis"`
	Syndrome           string             `json:"syndrome"`
	Items              []PrescriptionItem `json:"items"`
	TotalDosage        float64            `json:"total_dosage"`
	Status             PrescriptionStatus `json:"status"`
	StatusReason       string             `json:"status_reason,omitempty"`
	DispensingMode     string             `json:"dispensing_mode"`
	DecoctionMachineID string             `json:"decoction_machine_id,omitempty"`
	PotCount           int                `json:"pot_count,omitempty"`
	EstimatedFinish    *time.Time         `json:"estimated_finish,omitempty"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

type Herb struct {
	Name        string    `json:"name"`
	Stock       float64   `json:"stock"`
	ExpiryDate  time.Time `json:"expiry_date"`
	Price       float64   `json:"price"`
}

type HerbInventory struct {
	Herbs map[string]Herb
}

type CallNumberMessage struct {
	Type       string `json:"type"`
	SerialNum  string `json:"serial_num"`
	DoctorName string `json:"doctor_name"`
}
