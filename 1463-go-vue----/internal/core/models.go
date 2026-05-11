package core

import (
	"sync"
	"time"

	"safetymanager/internal/api"
)

type Zone struct {
	ID       string
	Name     string
	ParentID *string
	Children []*Zone
	mu       sync.RWMutex
}

type InspectionPlan struct {
	ID           string
	Name         string
	ZoneID       string
	InspectorIDs []string
	Items        []string
	Frequency    api.Frequency
	CreatedAt    time.Time
	mu           sync.RWMutex
}

type InspectionTaskItem struct {
	Index       int
	Name        string
	Result      *api.InspectionItemResult
	Description *string
}

type InspectionTask struct {
	ID           string
	PlanID       string
	ZoneID       string
	InspectorID  string
	Items        []InspectionTaskItem
	Status       api.TaskStatus
	ScheduledFor time.Time
	CompletedAt  *time.Time
	mu           sync.RWMutex
}

type LevelChangeRequest struct {
	ID            string
	HazardID      string
	ProposedLevel api.HazardLevel
	Reason        string
	Status        api.LevelChangeRequestStatus
	CreatedAt     time.Time
}

type Hazard struct {
	ID              string
	TaskID          string
	ZoneID          string
	ItemName        string
	Description     string
	Level           api.HazardLevel
	Status          api.HazardStatus
	ResponsibleDept string
	DueAt           time.Time
	RemediationNote *string
	RemediatedAt    *time.Time
	ClosedAt        *time.Time
	CreatedAt       time.Time
	Escalated       bool
	LevelChangeReq  *LevelChangeRequest
	mu              sync.RWMutex
}

type AuditLog struct {
	ID          string
	Operation   string
	EntityType  string
	EntityID    string
	Details     string
	PerformedAt time.Time
}

type Store struct {
	zones              map[string]*Zone
	plans              map[string]*InspectionPlan
	tasks              map[string]*InspectionTask
	hazards            map[string]*Hazard
	auditLogs          []AuditLog
	zoneMu             sync.RWMutex
	planMu             sync.RWMutex
	taskMu             sync.RWMutex
	hazardMu           sync.RWMutex
	auditMu            sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		zones:     make(map[string]*Zone),
		plans:     make(map[string]*InspectionPlan),
		tasks:     make(map[string]*InspectionTask),
		hazards:   make(map[string]*Hazard),
		auditLogs: make([]AuditLog, 0),
	}
}

func generateID() string {
	return time.Now().Format("20060102150405.000000000")
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func stringPtr(s string) *string {
	return &s
}

func toTimePtrPtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}

func toItemResultPtr(r *api.InspectionItemResult) *api.InspectionItemResult {
	if r == nil {
		return nil
	}
	return r
}

func toStringPtrPtr(s *string) *string {
	if s == nil {
		return nil
	}
	return s
}
