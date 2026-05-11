package core

import (
	"errors"
	"time"

	"safetymanager/internal/api"
)

func (s *Store) CreateZone(name string, parentID *string) (*Zone, error) {
	if name == "" {
		return nil, errors.New("zone name cannot be empty")
	}

	if parentID != nil {
		s.zoneMu.RLock()
		_, exists := s.zones[*parentID]
		s.zoneMu.RUnlock()
		if !exists {
			return nil, errors.New("parent zone not found")
		}
	}

	zone := &Zone{
		ID:       generateID(),
		Name:     name,
		ParentID: parentID,
		Children: make([]*Zone, 0),
	}

	s.zoneMu.Lock()
	s.zones[zone.ID] = zone
	if parentID != nil {
		parent := s.zones[*parentID]
		parent.Children = append(parent.Children, zone)
	}
	s.zoneMu.Unlock()

	s.appendAuditLog("zone_created", "zone", zone.ID, "zone="+name)

	return zone, nil
}

func (s *Store) GetZone(id string) (*Zone, error) {
	s.zoneMu.RLock()
	defer s.zoneMu.RUnlock()

	zone, exists := s.zones[id]
	if !exists {
		return nil, errors.New("zone not found")
	}
	return zone, nil
}

func (s *Store) GetAllZones() []*Zone {
	s.zoneMu.RLock()
	defer s.zoneMu.RUnlock()

	roots := make([]*Zone, 0)
	for _, zone := range s.zones {
		if zone.ParentID == nil {
			roots = append(roots, zone)
		}
	}
	return roots
}

func (s *Store) GetZoneAndChildren(zoneID string) ([]string, error) {
	s.zoneMu.RLock()
	defer s.zoneMu.RUnlock()

	zone, exists := s.zones[zoneID]
	if !exists {
		return nil, errors.New("zone not found")
	}

	ids := make([]string, 0)
	var collect func(z *Zone)
	collect = func(z *Zone) {
		ids = append(ids, z.ID)
		for _, child := range z.Children {
			collect(child)
		}
	}
	collect(zone)

	return ids, nil
}

func (s *Store) appendAuditLog(operation, entityType, entityID, details string) {
	log := AuditLog{
		ID:          generateID(),
		Operation:   operation,
		EntityType:  entityType,
		EntityID:    entityID,
		Details:     details,
		PerformedAt: time.Now(),
	}

	s.auditMu.Lock()
	s.auditLogs = append(s.auditLogs, log)
	s.auditMu.Unlock()
}

func buildZoneResponse(zone *Zone) api.ZoneResponse {
	children := make([]api.ZoneResponse, 0, len(zone.Children))
	for _, child := range zone.Children {
		children = append(children, buildZoneResponse(child))
	}

	return api.ZoneResponse{
		ID:       zone.ID,
		Name:     zone.Name,
		ParentID: zone.ParentID,
		Children: children,
	}
}
