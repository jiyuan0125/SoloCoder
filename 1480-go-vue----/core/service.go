package core

import (
	"vehicle-inspection/common"
)

type InspectionService struct {
	VehicleManager     *VehicleManager
	StationManager     *StationManager
	AppointmentManager *AppointmentManager
	InspectionManager *InspectionManager
}

func NewInspectionService() *InspectionService {
	vm := NewVehicleManager()
	sm := NewStationManager()
	am := NewAppointmentManager(vm, sm)
	im := NewInspectionManager(am, vm)
	
	return &InspectionService{
		VehicleManager:     vm,
		StationManager:     sm,
		AppointmentManager: am,
		InspectionManager: im,
	}
}

func (s *InspectionService) RegisterVehicle(req *common.RegisterVehicleRequest) (*common.Vehicle, error) {
	return s.VehicleManager.RegisterVehicle(req)
}

func (s *InspectionService) GetVehicleInfo(plateNumber string) (*common.VehicleInfoResponse, error) {
	return s.VehicleManager.GetVehicleInfo(plateNumber)
}

func (s *InspectionService) ListVehicles() []*common.Vehicle {
	return s.VehicleManager.ListVehicles()
}

func (s *InspectionService) CreateStation(req *common.CreateStationRequest) (*common.Station, error) {
	return s.StationManager.CreateStation(req)
}

func (s *InspectionService) AddSchedule(req *common.AddScheduleRequest) (*common.Schedule, error) {
	return s.StationManager.AddSchedule(req)
}

func (s *InspectionService) ListStations() []*common.Station {
	return s.StationManager.ListStations()
}

func (s *InspectionService) CreateAppointment(req *common.CreateAppointmentRequest) (*common.Appointment, error) {
	return s.AppointmentManager.CreateAppointment(req)
}

func (s *InspectionService) GetAppointment(id string) (*common.Appointment, error) {
	return s.AppointmentManager.GetAppointment(id)
}

func (s *InspectionService) ListAppointments() []*common.Appointment {
	return s.AppointmentManager.ListAppointments()
}

func (s *InspectionService) StartInspection(req *common.StartInspectionRequest) (*common.InspectionProcess, error) {
	return s.InspectionManager.StartInspection(req)
}

func (s *InspectionService) CompleteStep(req *common.CompleteStepRequest) (*common.InspectionProcess, error) {
	return s.InspectionManager.CompleteStep(req)
}

func (s *InspectionService) StartRecheck(req *common.StartRecheckRequest) (*common.InspectionProcess, error) {
	return s.InspectionManager.StartRecheck(req)
}

func (s *InspectionService) GetProcess(id string) (*common.InspectionProcess, error) {
	return s.InspectionManager.GetProcess(id)
}

func (s *InspectionService) GetReport(id string) (*common.InspectionReport, error) {
	return s.InspectionManager.GetReport(id)
}

func (s *InspectionService) ListProcesses() []*common.InspectionProcess {
	return s.InspectionManager.ListProcesses()
}
