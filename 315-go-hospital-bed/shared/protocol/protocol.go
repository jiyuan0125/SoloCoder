package protocol

type BedType string

const (
	BedTypeSingle   BedType = "single"
	BedTypeDouble   BedType = "double"
	BedTypeTriple   BedType = "triple"
	BedTypeExtra    BedType = "extra"
)

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type WardConfig struct {
	WardName     string `json:"ward_name"`
	SingleBeds   int    `json:"single_beds"`
	DoubleBeds   int    `json:"double_beds"`
	TripleBeds   int    `json:"triple_beds"`
	ExtraBedLimit int   `json:"extra_bed_limit"`
}

type DepartmentConfigRequest struct {
	DepartmentName string       `json:"department_name"`
	Wards          []WardConfig `json:"wards"`
}

type AdmissionRequest struct {
	PatientID      string  `json:"patient_id"`
	PatientName    string  `json:"patient_name"`
	DepartmentName string  `json:"department_name"`
	WardName       *string `json:"ward_name,omitempty"`
	PreferredType  BedType `json:"preferred_type"`
}

type AdmissionResult int

const (
	AdmissionResultAllocated AdmissionResult = iota
	AdmissionResultQueued
	AdmissionResultFailed
)

type DowngradeReason struct {
	PreferredType BedType `json:"preferred_type"`
	ActualType    BedType `json:"actual_type"`
	Reason        string  `json:"reason"`
}

type AdmissionResponse struct {
	Response
	Result          AdmissionResult `json:"result"`
	BedID           string          `json:"bed_id,omitempty"`
	BedType         BedType         `json:"bed_type,omitempty"`
	QueuePosition   int             `json:"queue_position,omitempty"`
	DowngradeReason *DowngradeReason `json:"downgrade_reason,omitempty"`
}

type DischargeRequest struct {
	PatientID string `json:"patient_id"`
}

type DischargeResponse struct {
	Response
	BedID string `json:"bed_id,omitempty"`
}

type BedStatus struct {
	WardName    string `json:"ward_name"`
	BedType     BedType `json:"bed_type"`
	Total       int    `json:"total"`
	Available   int    `json:"available"`
}

type DepartmentStatusResponse struct {
	Response
	DepartmentName string       `json:"department_name,omitempty"`
	BedStatuses    []BedStatus  `json:"bed_statuses,omitempty"`
	QueueLength    int          `json:"queue_length,omitempty"`
}

type PatientBedRequest struct {
	PatientID string `json:"patient_id"`
}

type PatientBedResponse struct {
	Response
	PatientID       string    `json:"patient_id,omitempty"`
	PatientName     string    `json:"patient_name,omitempty"`
	BedID           string    `json:"bed_id,omitempty"`
	DepartmentName  string    `json:"department_name,omitempty"`
	WardName        string    `json:"ward_name,omitempty"`
	BedType         BedType   `json:"bed_type,omitempty"`
	InQueue         bool      `json:"in_queue,omitempty"`
	QueuePosition   int       `json:"queue_position,omitempty"`
}

type QueueStatusResponse struct {
	Response
	DepartmentName string       `json:"department_name,omitempty"`
	QueueLength    int          `json:"queue_length,omitempty"`
	QueueItems     []QueueItem  `json:"queue_items,omitempty"`
}

type QueueItem struct {
	PatientID     string  `json:"patient_id"`
	PatientName   string  `json:"patient_name"`
	PreferredType BedType `json:"preferred_type"`
	Position      int     `json:"position"`
}
