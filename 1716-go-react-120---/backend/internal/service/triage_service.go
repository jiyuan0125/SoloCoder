package service

import (
	"errors"
	"time"

	"ambulance-scheduler/internal/models"
	"ambulance-scheduler/internal/store"
	"ambulance-scheduler/pkg/distance"
)

type Hospital struct {
	Name     string
	Location string
	Capabilities []string
}

var hospitals = []Hospital{
	{Name: "市中心医院", Location: "市中心", Capabilities: []string{"心内科", "神经内科", "急诊科", "ICU"}},
	{Name: "东区人民医院", Location: "东城区", Capabilities: []string{"急诊科", "创伤外科", "儿科"}},
	{Name: "西区中医院", Location: "西城区", Capabilities: []string{"急诊科", "心内科", "中医骨伤科"}},
	{Name: "南区妇女儿童医院", Location: "南城区", Capabilities: []string{"急诊科", "妇产科", "儿科", "新生儿科"}},
	{Name: "北区综合医院", Location: "北城区", Capabilities: []string{"急诊科", "神经内科", "骨科"}},
}

type TriageService struct {
	store *store.Store
}

func NewTriageService(store *store.Store) *TriageService {
	return &TriageService{store: store}
}

type CreateTriageRequest struct {
	CallID               string `json:"call_id"`
	DispatchID           string `json:"dispatch_id"`
	HeartRate            int    `json:"heart_rate"`
	BloodPressure        string `json:"blood_pressure"`
	OxygenSaturation     float64 `json:"oxygen_saturation"`
	PreliminaryDiagnosis string `json:"preliminary_diagnosis"`
	RescueMeasures       string `json:"rescue_measures,omitempty"`
	RescueDuration       int    `json:"rescue_duration_minutes,omitempty"`
}

func (s *TriageService) CreateTriageRecord(req *CreateTriageRequest) (*models.TriageRecord, error) {
	call, ok := s.store.GetCall(req.CallID)
	if !ok {
		return nil, errors.New("求救记录不存在")
	}

	dispatch, ok := s.store.GetDispatchRecord(req.DispatchID)
	if !ok {
		return nil, errors.New("出车记录不存在")
	}

	targetHospital := s.findBestHospital(call.Location, req.PreliminaryDiagnosis)

	triage := models.NewTriageRecord()
	triage.CallID = req.CallID
	triage.DispatchID = req.DispatchID
	triage.HeartRate = req.HeartRate
	triage.BloodPressure = req.BloodPressure
	triage.OxygenSaturation = req.OxygenSaturation
	triage.PreliminaryDiagnosis = req.PreliminaryDiagnosis
	triage.TargetHospital = targetHospital
	triage.RescueMeasures = req.RescueMeasures
	triage.RescueDuration = req.RescueDuration

	s.store.CreateTriageRecord(triage)

	now := time.Now()
	arrivalTime := now
	dispatch.ArrivalTime = &arrivalTime
	s.store.CreateDispatchRecord(dispatch)

	return triage, nil
}

func (s *TriageService) findBestHospital(callLocation, diagnosis string) string {
	callArea := distance.DetectArea(callLocation)
	
	var candidates []Hospital
	for _, h := range hospitals {
		if s.hasCapability(&h, diagnosis) {
			candidates = append(candidates, h)
		}
	}

	if len(candidates) == 0 {
		candidates = hospitals
	}

	best := candidates[0]
	bestDist := distance.CalculateDistance(callArea, best.Location)

	for _, h := range candidates[1:] {
		dist := distance.CalculateDistance(callArea, h.Location)
		if dist < bestDist {
			best = h
			bestDist = dist
		}
	}

	return best.Name
}

func (s *TriageService) hasCapability(hospital *Hospital, diagnosis string) bool {
	for _, cap := range hospital.Capabilities {
		if matchesCapability(cap, diagnosis) {
			return true
		}
	}
	return false
}

func matchesCapability(capability, diagnosis string) bool {
	mapping := map[string][]string{
		"心内科":     {"心肌梗死", "心脏病", "心绞痛", "胸痛", "心律失常"},
		"神经内科":   {"脑卒中", "脑出血", "脑梗死", "中风", "昏迷"},
		"急诊科":     {"外伤", "出血", "呼吸困难", "中毒"},
		"ICU":      {"重症", "呼吸衰竭", "多器官衰竭"},
		"创伤外科":   {"骨折", "外伤", "车祸", "坠落"},
		"儿科":      {"儿童", "小儿"},
		"新生儿科":   {"新生儿", "早产"},
		"中医骨伤科": {"骨伤", "骨折", "脱臼"},
		"妇产科":     {"临产", "妊娠", "分娩", "流产"},
		"骨科":      {"骨折", "脱臼", "脊柱损伤"},
	}

	if keywords, ok := mapping[capability]; ok {
		for _, kw := range keywords {
			if contains(diagnosis, kw) {
				return true
			}
		}
	}

	return true
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func (s *TriageService) ListTriageRecords() []*models.TriageRecord {
	return s.store.ListTriageRecords()
}
