package core

import (
	"errors"

	"safetymanager/internal/api"
)

type Service struct {
	store *Store
}

func NewService() *Service {
	return &Service{
		store: NewStore(),
	}
}

func (svc *Service) CreateZone(name string, parentID *string) (*api.ZoneResponse, error) {
	zone, err := svc.store.CreateZone(name, parentID)
	if err != nil {
		return nil, err
	}
	resp := buildZoneResponse(zone)
	return &resp, nil
}

func (svc *Service) GetAllZones() ([]api.ZoneResponse, error) {
	zones := svc.store.GetAllZones()
	resp := make([]api.ZoneResponse, 0, len(zones))
	for _, zone := range zones {
		resp = append(resp, buildZoneResponse(zone))
	}
	return resp, nil
}

func (svc *Service) CreateInspectionPlan(req *api.CreateInspectionPlanRequest) (*api.InspectionPlanResponse, error) {
	plan, err := svc.store.CreateInspectionPlan(
		req.Name,
		req.ZoneID,
		req.InspectorIDs,
		req.Items,
		req.Frequency,
	)
	if err != nil {
		return nil, err
	}

	zone, err := svc.store.GetZone(plan.ZoneID)
	if err != nil {
		return nil, err
	}

	resp := buildPlanResponse(plan, zone.Name)
	return &resp, nil
}

func (svc *Service) GetAllInspectionPlans() ([]api.InspectionPlanResponse, error) {
	plans := svc.store.GetAllInspectionPlans()
	resp := make([]api.InspectionPlanResponse, 0, len(plans))
	for _, plan := range plans {
		zone, err := svc.store.GetZone(plan.ZoneID)
		if err != nil {
			return nil, err
		}
		resp = append(resp, buildPlanResponse(plan, zone.Name))
	}
	return resp, nil
}

func (svc *Service) GenerateTasks() error {
	return svc.store.GenerateTasks()
}

func (svc *Service) GetAllInspectionTasks() ([]api.InspectionTaskResponse, error) {
	tasks := svc.store.GetAllInspectionTasks()
	resp := make([]api.InspectionTaskResponse, 0, len(tasks))
	for _, task := range tasks {
		plan, err := svc.store.GetInspectionPlan(task.PlanID)
		if err != nil {
			return nil, err
		}
		zone, err := svc.store.GetZone(task.ZoneID)
		if err != nil {
			return nil, err
		}
		resp = append(resp, buildTaskResponse(task, plan.Name, zone.Name))
	}
	return resp, nil
}

func (svc *Service) GetInspectionTask(taskID string) (*api.InspectionTaskResponse, error) {
	task, err := svc.store.GetInspectionTask(taskID)
	if err != nil {
		return nil, err
	}
	plan, err := svc.store.GetInspectionPlan(task.PlanID)
	if err != nil {
		return nil, err
	}
	zone, err := svc.store.GetZone(task.ZoneID)
	if err != nil {
		return nil, err
	}
	resp := buildTaskResponse(task, plan.Name, zone.Name)
	return &resp, nil
}

func (svc *Service) SubmitInspection(req *api.SubmitInspectionRequest) ([]api.HazardResponse, error) {
	hazards, err := svc.store.SubmitInspection(req.TaskID, req.Items)
	if err != nil {
		return nil, err
	}

	resp := make([]api.HazardResponse, 0, len(hazards))
	for _, hazard := range hazards {
		zone, err := svc.store.GetZone(hazard.ZoneID)
		if err != nil {
			return nil, err
		}
		resp = append(resp, buildHazardResponse(hazard, zone.Name))
	}
	return resp, nil
}

func (svc *Service) GetAllHazards() ([]api.HazardResponse, error) {
	hazards := svc.store.GetAllHazards()
	resp := make([]api.HazardResponse, 0, len(hazards))
	for _, hazard := range hazards {
		zone, err := svc.store.GetZone(hazard.ZoneID)
		if err != nil {
			return nil, err
		}
		resp = append(resp, buildHazardResponse(hazard, zone.Name))
	}
	return resp, nil
}

func (svc *Service) GetHazard(hazardID string) (*api.HazardResponse, error) {
	hazard, err := svc.store.GetHazard(hazardID)
	if err != nil {
		return nil, err
	}
	zone, err := svc.store.GetZone(hazard.ZoneID)
	if err != nil {
		return nil, err
	}
	resp := buildHazardResponse(hazard, zone.Name)
	return &resp, nil
}

func (svc *Service) SubmitRemediation(req *api.SubmitRemediationRequest) error {
	return svc.store.SubmitRemediation(req.HazardID, req.Note)
}

func (svc *Service) ReviewRemediation(req *api.ReviewRemediationRequest) error {
	return svc.store.ReviewRemediation(req.HazardID, req.Approved)
}

func (svc *Service) RequestLevelChange(req *api.RequestLevelChangeRequest) (*api.LevelChangeRequestResponse, error) {
	levelReq, err := svc.store.RequestLevelChange(req.HazardID, req.ProposedLevel, req.Reason)
	if err != nil {
		return nil, err
	}
	resp := buildLevelChangeResponse(levelReq)
	return &resp, nil
}

func (svc *Service) ReviewLevelChange(req *api.ReviewLevelChangeRequest) error {
	return svc.store.ReviewLevelChange(req.RequestID, req.Approved)
}

func (svc *Service) GetLevelChangeRequest(requestID string) (*api.LevelChangeRequestResponse, error) {
	hazards := svc.store.GetAllHazards()
	for _, hazard := range hazards {
		if hazard.LevelChangeReq != nil && hazard.LevelChangeReq.ID == requestID {
			resp := buildLevelChangeResponse(hazard.LevelChangeReq)
			return &resp, nil
		}
	}
	return nil, errors.New("level change request not found")
}

func (svc *Service) CheckEscalations() []string {
	return svc.store.CheckEscalations()
}

func (svc *Service) GetAllAuditLogs() []api.AuditLogResponse {
	logs := svc.store.GetAllAuditLogs()
	resp := make([]api.AuditLogResponse, 0, len(logs))
	for _, log := range logs {
		resp = append(resp, buildAuditLogResponse(log))
	}
	return resp
}

func (svc *Service) GetAuditLogsByEntity(entityType, entityID string) []api.AuditLogResponse {
	logs := svc.store.GetAuditLogsByEntity(entityType, entityID)
	resp := make([]api.AuditLogResponse, 0, len(logs))
	for _, log := range logs {
		resp = append(resp, buildAuditLogResponse(log))
	}
	return resp
}

func (svc *Service) ExportCriticalHazardLogs() []api.AuditLogResponse {
	logs := svc.store.ExportCriticalHazardLogs()
	resp := make([]api.AuditLogResponse, 0, len(logs))
	for _, log := range logs {
		resp = append(resp, buildAuditLogResponse(log))
	}
	return resp
}
