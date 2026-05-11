package core

import (
	"cleaning-service/internal/common"
	"fmt"
	"time"
)

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{store: store}
}

func (s *Service) CreateClient(req *common.CreateClientRequest) (*common.Client, error) {
	client := &common.Client{
		ID:        s.store.nextID("C"),
		Name:      req.Name,
		CreatedAt: time.Now(),
	}

	contract := &common.Contract{
		ID:             s.store.nextID("CT"),
		ClientID:       client.ID,
		ServiceAddress: req.ServiceAddress,
		Area:           req.Area,
		Frequency:      req.Frequency,
		SpecialNotes:   req.SpecialNotes,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
	}

	client.Contract = contract

	s.store.SaveContract(contract)
	s.store.SaveClient(client)

	return client, nil
}

func (s *Service) GetClient(id string) (*common.Client, error) {
	client := s.store.GetClient(id)
	if client == nil {
		return nil, fmt.Errorf("client not found")
	}
	return client, nil
}

func (s *Service) ListClients() []*common.Client {
	return s.store.ListClients()
}

func (s *Service) CreateServiceArea(req *common.CreateServiceAreaRequest) (*common.ServiceArea, error) {
	if s.store.GetClient(req.ClientID) == nil {
		return nil, fmt.Errorf("client not found")
	}

	area := &common.ServiceArea{
		ID:       s.store.nextID("SA"),
		Name:     req.Name,
		Address:  req.Address,
		Area:     req.Area,
		ClientID: req.ClientID,
		ZoneID:   req.ZoneID,
	}

	s.store.SaveServiceArea(area)

	client := s.store.GetClient(req.ClientID)
	if client.Contract != nil {
		client.Contract.ServiceAreas = append(client.Contract.ServiceAreas, area.ID)
		s.store.SaveContract(client.Contract)
	}

	return area, nil
}

func (s *Service) ListServiceAreas() []*common.ServiceArea {
	return s.store.ListServiceAreas()
}

func (s *Service) CreateZone(req *common.CreateZoneRequest) (*common.Zone, error) {
	zone := &common.Zone{
		ID:   s.store.nextID("Z"),
		Name: req.Name,
	}

	s.store.SaveZone(zone)
	return zone, nil
}

func (s *Service) ListZones() []*common.Zone {
	return s.store.ListZones()
}

func (s *Service) CreateTeam(req *common.CreateTeamRequest) (*common.Team, error) {
	if s.store.GetZone(req.ZoneID) == nil {
		return nil, fmt.Errorf("zone not found")
	}

	team := &common.Team{
		ID:     s.store.nextID("T"),
		Name:   req.Name,
		ZoneID: req.ZoneID,
	}

	s.store.SaveTeam(team)
	return team, nil
}

func (s *Service) ListTeams() []*common.Team {
	return s.store.ListTeams()
}

func (s *Service) CreateCleaner(req *common.CreateCleanerRequest) (*common.Cleaner, error) {
	if s.store.GetTeam(req.TeamID) == nil {
		return nil, fmt.Errorf("team not found")
	}

	cleaner := &common.Cleaner{
		ID:        s.store.nextID("CL"),
		Name:      req.Name,
		Skill:     req.Skill,
		TeamID:    req.TeamID,
		CreatedAt: time.Now(),
	}

	s.store.SaveCleaner(cleaner)
	return cleaner, nil
}

func (s *Service) ListCleaners() []*common.Cleaner {
	return s.store.ListCleaners()
}

func (s *Service) CreateSchedule(req *common.CreateScheduleRequest) (*common.WeeklySchedule, error) {
	if s.store.GetTeam(req.TeamID) == nil {
		return nil, fmt.Errorf("team not found")
	}

	daySchedules := make(map[time.Weekday][]common.Shift)
	for day, shifts := range req.Days {
		daySchedules[day] = make([]common.Shift, 0, len(shifts))
		for _, sr := range shifts {
			daySchedules[day] = append(daySchedules[day], common.Shift{StartHour: sr.StartHour, EndHour: sr.EndHour})
		}
	}

	schedule := &common.WeeklySchedule{
		ID:         s.store.nextID("SC"),
		TeamID:     req.TeamID,
		WeekStart:  req.WeekStart,
		WeekEnd:    req.WeekEnd,
		DaySchedules: daySchedules,
	}

	s.store.SaveSchedule(schedule)
	return schedule, nil
}

func (s *Service) CreateLeaveRequest(req *common.CreateLeaveRequest) (*common.LeaveRequest, error) {
	if s.store.GetCleaner(req.CleanerID) == nil {
		return nil, fmt.Errorf("cleaner not found")
	}

	now := time.Now()
	if req.Date.Before(now.Add(24 * time.Hour)) {
		return nil, fmt.Errorf("leave request must be made at least one day in advance")
	}

	leave := &common.LeaveRequest{
		ID:          s.store.nextID("LV"),
		CleanerID:   req.CleanerID,
		Date:        req.Date,
		Reason:      req.Reason,
		Status:      common.LeaveStatusPending,
		RequestedAt: now,
	}

	s.store.SaveLeave(leave)
	return leave, nil
}

func (s *Service) ApproveLeave(leaveID string, approved bool) (*common.LeaveRequest, error) {
	leave := s.store.GetLeave(leaveID)
	if leave == nil {
		return nil, fmt.Errorf("leave request not found")
	}

	if approved {
		leave.Status = common.LeaveStatusApproved
		now := time.Now()
		leave.ApprovedAt = &now
	} else {
		leave.Status = common.LeaveStatusRejected
	}

	s.store.SaveLeave(leave)
	return leave, nil
}

func (s *Service) ListLeaves() []*common.LeaveRequest {
	return s.store.ListLeaves()
}
