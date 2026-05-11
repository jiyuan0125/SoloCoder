package core

import (
	"time"

	"safetymanager/internal/api"
)

func (s *Store) GetAllAuditLogs() []AuditLog {
	s.auditMu.RLock()
	defer s.auditMu.RUnlock()

	logs := make([]AuditLog, len(s.auditLogs))
	copy(logs, s.auditLogs)
	return logs
}

func (s *Store) GetAuditLogsByEntity(entityType, entityID string) []AuditLog {
	s.auditMu.RLock()
	defer s.auditMu.RUnlock()

	logs := make([]AuditLog, 0)
	for _, log := range s.auditLogs {
		if log.EntityType == entityType && log.EntityID == entityID {
			logs = append(logs, log)
		}
	}
	return logs
}

func (s *Store) GetAuditLogsByOperation(operation string) []AuditLog {
	s.auditMu.RLock()
	defer s.auditMu.RUnlock()

	logs := make([]AuditLog, 0)
	for _, log := range s.auditLogs {
		if log.Operation == operation {
			logs = append(logs, log)
		}
	}
	return logs
}

func (s *Store) ExportCriticalHazardLogs() []AuditLog {
	s.auditMu.RLock()
	defer s.auditMu.RUnlock()

	s.hazardMu.RLock()
	defer s.hazardMu.RUnlock()

	criticalHazardIDs := make(map[string]bool)
	for _, hazard := range s.hazards {
		if hazard.Level == api.HazardLevelCritical {
			criticalHazardIDs[hazard.ID] = true
		}
	}

	logs := make([]AuditLog, 0)
	for _, log := range s.auditLogs {
		if log.EntityType == "hazard" && criticalHazardIDs[log.EntityID] {
			logs = append(logs, log)
		}
	}
	return logs
}

func buildAuditLogResponse(log AuditLog) api.AuditLogResponse {
	return api.AuditLogResponse{
		ID:          log.ID,
		Operation:   log.Operation,
		EntityType:  log.EntityType,
		EntityID:    log.EntityID,
		Details:     log.Details,
		PerformedAt: log.PerformedAt.Format(time.RFC3339),
	}
}
