package core

type Service struct {
	store             *Store
	DeviceService     *DeviceService
	InspectionService *InspectionService
	DrillService      *DrillService
	Scheduler         *Scheduler
}

func NewService() *Service {
	store := NewStore()

	ds := NewDeviceService(store)
	is := NewInspectionService(store)
	drs := NewDrillService(store)
	sched := NewScheduler(ds, is, drs)
	sched.RegisterDefaultTasks()

	return &Service{
		store:             store,
		DeviceService:     ds,
		InspectionService: is,
		DrillService:      drs,
		Scheduler:         sched,
	}
}

func (s *Service) Start() {
	s.Scheduler.Start()
}

func (s *Service) Stop() {
	s.Scheduler.Stop()
}
