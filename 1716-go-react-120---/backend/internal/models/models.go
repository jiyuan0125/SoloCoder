package models

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type SeverityLevel string

const (
	SeverityLevel1 SeverityLevel = "一级濒危"
	SeverityLevel2 SeverityLevel = "二级危重"
	SeverityLevel3 SeverityLevel = "三级急症"
	SeverityLevel4 SeverityLevel = "四级非急症"
)

type AgeGroup string

const (
	AgeGroupChild    AgeGroup = "儿童"
	AgeGroupAdult    AgeGroup = "成人"
	AgeGroupElderly  AgeGroup = "老人"
)

type VehicleType string

const (
	VehicleTypeNormal     VehicleType = "普通转运车"
	VehicleTypeEmergency  VehicleType = "抢救监护车"
	VehicleTypeNeonate    VehicleType = "新生儿转运车"
)

type VehicleStatus string

const (
	VehicleStatusIdle       VehicleStatus = "空闲"
	VehicleStatusInTransit  VehicleStatus = "出车中"
	VehicleStatusReturning  VehicleStatus = "返回途中"
	VehicleStatusMaintenance VehicleStatus = "维护中"
)

type CallStatus string

const (
	CallStatusPendingAccept CallStatus = "待接单"
	CallStatusAccepted      CallStatus = "已接单"
	CallStatusProcessing    CallStatus = "处理中"
	CallStatusPendingCheck  CallStatus = "待验收"
	CallStatusCompleted     CallStatus = "已完成"
	CallStatusClosed        CallStatus = "已关闭"
)

type EmergencyCall struct {
	ID             string        `json:"id"`
	CallTime       time.Time     `json:"call_time"`
	CallerName     string        `json:"caller_name"`
	CallerPhone    string        `json:"caller_phone"`
	Location       string        `json:"location"`
	PatientCount   int           `json:"patient_count"`
	PatientGender  string        `json:"patient_gender"`
	PatientAgeGroup AgeGroup    `json:"patient_age_group"`
	ChiefComplaint string        `json:"chief_complaint"`
	SeverityLevel  SeverityLevel `json:"severity_level"`
	Status         CallStatus    `json:"status"`
	AssignedVehicleID string     `json:"assigned_vehicle_id,omitempty"`
	DispatchTime   *time.Time    `json:"dispatch_time,omitempty"`
	CreateTime     time.Time     `json:"create_time"`
	UpdateTime     time.Time     `json:"update_time"`
	InQueue        bool          `json:"in_queue"`
	QueueTime      *time.Time    `json:"queue_time,omitempty"`
	LastUpgradeTime *time.Time   `json:"last_upgrade_time,omitempty"`
	BillID         string        `json:"bill_id,omitempty"`
	Amount         float64       `json:"amount,omitempty"`
}

type Ambulance struct {
	ID              string        `json:"id"`
	VehicleNumber   string        `json:"vehicle_number"`
	VehicleType     VehicleType   `json:"vehicle_type"`
	CurrentStatus   VehicleStatus `json:"current_status"`
	CurrentLocation string        `json:"current_location"`
	DoctorCount     int           `json:"doctor_count"`
	NurseCount      int           `json:"nurse_count"`
	PatientCount    int           `json:"patient_count"`
	CurrentStatusHistory []StatusHistory `json:"status_history"`
	CreateTime      time.Time     `json:"create_time"`
	UpdateTime      time.Time     `json:"update_time"`
}

type StatusHistory struct {
	Status  VehicleStatus `json:"status"`
	Time    time.Time     `json:"time"`
}

type DispatchRecord struct {
	ID                  string    `json:"id"`
	CallID              string    `json:"call_id"`
	VehicleID           string    `json:"vehicle_id"`
	DispatchTime        time.Time `json:"dispatch_time"`
	TargetLocation      string    `json:"target_location"`
	EstimatedArrivalTime int      `json:"estimated_arrival_time_minutes"`
	EstimatedDistance   float64   `json:"estimated_distance_km"`
	ArrivalTime         *time.Time `json:"arrival_time,omitempty"`
	StartReturnTime     *time.Time `json:"start_return_time,omitempty"`
	BackAtStationTime   *time.Time `json:"back_at_station_time,omitempty"`
	CreateTime          time.Time `json:"create_time"`
}

type TriageRecord struct {
	ID              string    `json:"id"`
	CallID          string    `json:"call_id"`
	DispatchID      string    `json:"dispatch_id"`
	HeartRate       int       `json:"heart_rate"`
	BloodPressure   string    `json:"blood_pressure"`
	OxygenSaturation float64  `json:"oxygen_saturation"`
	PreliminaryDiagnosis string `json:"preliminary_diagnosis"`
	TargetHospital  string    `json:"target_hospital"`
	RescueMeasures  string    `json:"rescue_measures,omitempty"`
	RescueDuration  int       `json:"rescue_duration_minutes,omitempty"`
	CreateTime      time.Time `json:"create_time"`
}

type Statistics struct {
	TodayCallsCount         int     `json:"today_calls_count"`
	TotalCallsCount         int     `json:"total_calls_count"`
	AverageResponseTime     float64 `json:"average_response_time_minutes"`
	VehicleUtilizationRate  float64 `json:"vehicle_utilization_rate"`
	IdleVehiclesCount       int     `json:"idle_vehicles_count"`
	InTransitVehiclesCount  int     `json:"in_transit_vehicles_count"`
	MaintenanceVehiclesCount int    `json:"maintenance_vehicles_count"`
	WaitingInQueueCount     int     `json:"waiting_in_queue_count"`
	Level12CallsToday       int     `json:"level12_calls_today"`
	Level34CallsToday       int     `json:"level34_calls_today"`
}

type Bill struct {
	ID      string  `json:"id"`
	Amount  float64 `json:"amount"`
	Paid    bool    `json:"paid"`
}

func NewEmergencyCall() *EmergencyCall {
	return &EmergencyCall{
		ID:         uuid.New().String(),
		Status:     CallStatusPendingAccept,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
}

func NewAmbulance() *Ambulance {
	now := time.Now()
	return &Ambulance{
		ID:              uuid.New().String(),
		CurrentStatus:   VehicleStatusIdle,
		CurrentStatusHistory: []StatusHistory{{Status: VehicleStatusIdle, Time: now}},
		CreateTime:      now,
		UpdateTime:      now,
	}
}

func NewDispatchRecord() *DispatchRecord {
	return &DispatchRecord{
		ID:         uuid.New().String(),
		CreateTime: time.Now(),
	}
}

func NewTriageRecord() *TriageRecord {
	return &TriageRecord{
		ID:         uuid.New().String(),
		CreateTime: time.Now(),
	}
}

func ValidSeverity(level SeverityLevel) bool {
	switch level {
	case SeverityLevel1, SeverityLevel2, SeverityLevel3, SeverityLevel4:
		return true
	default:
		return false
	}
}

func ParseSeverity(input string) (SeverityLevel, bool) {
	input = trimSpace(input)
	
	if ValidSeverity(SeverityLevel(input)) {
		return SeverityLevel(input), true
	}

	switch input {
	case "1", "一级", "一级濒危":
		return SeverityLevel1, true
	case "2", "二级", "二级危重":
		return SeverityLevel2, true
	case "3", "三级", "三级急症":
		return SeverityLevel3, true
	case "4", "四级", "四级非急症":
		return SeverityLevel4, true
	default:
		return "", false
	}
}

func trimSpace(s string) string {
	result := []rune{}
	for _, r := range s {
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
			result = append(result, r)
		}
	}
	return string(result)
}

func (s *SeverityLevel) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		parsed, ok := ParseSeverity(str)
		if ok {
			*s = parsed
			return nil
		}
	}

	var num int
	if err := json.Unmarshal(data, &num); err == nil {
		parsed, ok := ParseSeverity(strconv.Itoa(num))
		if ok {
			*s = parsed
			return nil
		}
	}

	return json.Unmarshal(data, (*string)(s))
}

func ValidVehicleStatusTransition(from, to VehicleStatus) bool {
	switch from {
	case VehicleStatusIdle:
		return to == VehicleStatusInTransit
	case VehicleStatusInTransit:
		return to == VehicleStatusReturning
	case VehicleStatusReturning:
		return to == VehicleStatusIdle
	case VehicleStatusMaintenance:
		return to == VehicleStatusIdle
	default:
		return false
	}
}
