package core

type AuditService struct {
	store *Store
}

func NewAuditService(store *Store) *AuditService {
	return &AuditService{store: store}
}

func (s *AuditService) GetAuditLogs(contractID string) []*AuditLog {
	return s.store.GetAuditLogs(contractID)
}

func (s *AuditService) GetAuditLogsByOperation(contractID string, operation string) []*AuditLog {
	logs := s.store.GetAuditLogs(contractID)
	filtered := []*AuditLog{}
	for _, l := range logs {
		if l.Operation == operation {
			filtered = append(filtered, l)
		}
	}
	return filtered
}

func (s *AuditService) GetAuditLogsByOperator(contractID string, operator string) []*AuditLog {
	logs := s.store.GetAuditLogs(contractID)
	filtered := []*AuditLog{}
	for _, l := range logs {
		if l.Operator == operator {
			filtered = append(filtered, l)
		}
	}
	return filtered
}
