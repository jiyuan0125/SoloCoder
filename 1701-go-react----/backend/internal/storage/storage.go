package storage

import (
	"cmms/internal/models"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Storage struct {
	patients       sync.Map
	registrations  sync.Map
	prescriptions  sync.Map
	herbs          sync.Map
	serialCounter  map[string]*uint64
	counterMu      sync.Mutex
}

func NewStorage() *Storage {
	s := &Storage{
		serialCounter: make(map[string]*uint64),
	}
	s.initHerbs()
	return s
}

func (s *Storage) initHerbs() {
	future1 := time.Now().AddDate(1, 0, 0)
	future2 := time.Now().AddDate(0, 6, 0)
	soon := time.Now().AddDate(0, 0, 20)
	expired := time.Now().AddDate(0, 0, -10)

	defaultHerbs := []models.Herb{
		{Name: "当归", Stock: 5000, ExpiryDate: future1, Price: 2.5},
		{Name: "黄芪", Stock: 8000, ExpiryDate: future1, Price: 1.8},
		{Name: "白术", Stock: 3000, ExpiryDate: future2, Price: 1.5},
		{Name: "茯苓", Stock: 4500, ExpiryDate: future2, Price: 1.2},
		{Name: "甘草", Stock: 6000, ExpiryDate: soon, Price: 0.8},
		{Name: "川芎", Stock: 2500, ExpiryDate: future1, Price: 2.0},
		{Name: "白芍", Stock: 3500, ExpiryDate: future2, Price: 1.6},
		{Name: "熟地", Stock: 4000, ExpiryDate: expired, Price: 2.2},
		{Name: "人参", Stock: 500, ExpiryDate: future1, Price: 15.0},
		{Name: "枸杞", Stock: 2000, ExpiryDate: future2, Price: 3.0},
		{Name: "菊花", Stock: 1500, ExpiryDate: future1, Price: 2.5},
		{Name: "金银花", Stock: 1800, ExpiryDate: future2, Price: 4.0},
		{Name: "薄荷", Stock: 1000, ExpiryDate: soon, Price: 1.5},
		{Name: "陈皮", Stock: 3000, ExpiryDate: future1, Price: 1.0},
		{Name: "半夏", Stock: 2000, ExpiryDate: future2, Price: 2.8},
	}

	for _, herb := range defaultHerbs {
		s.herbs.Store(herb.Name, herb)
	}
}

func (s *Storage) GetSerialNum() (string, error) {
	now := time.Now()
	dateStr := now.Format("20060102")

	s.counterMu.Lock()
	counter, exists := s.serialCounter[dateStr]
	if !exists {
		var initial uint64 = 0
		counter = &initial
		s.serialCounter[dateStr] = counter
	}
	s.counterMu.Unlock()

	newNum := atomic.AddUint64(counter, 1)
	return fmt.Sprintf("%s%03d", dateStr, newNum), nil
}

func (s *Storage) SavePatient(patient *models.Patient) {
	s.patients.Store(patient.ID, *patient)
}

func (s *Storage) GetPatient(id string) (models.Patient, bool) {
	if val, ok := s.patients.Load(id); ok {
		return val.(models.Patient), true
	}
	return models.Patient{}, false
}

func (s *Storage) SaveRegistration(reg *models.Registration) {
	s.registrations.Store(reg.ID, *reg)
}

func (s *Storage) GetRegistration(id string) (models.Registration, bool) {
	if val, ok := s.registrations.Load(id); ok {
		return val.(models.Registration), true
	}
	return models.Registration{}, false
}

func (s *Storage) GetTodayRegistrations() []models.Registration {
	today := time.Now().Format("20060102")
	var regs []models.Registration

	s.registrations.Range(func(key, value interface{}) bool {
		reg := value.(models.Registration)
		if reg.Date == today {
			regs = append(regs, reg)
		}
		return true
	})
	return regs
}

func (s *Storage) UpdateRegistrationStatus(id string, status models.PatientStatus) bool {
	if val, ok := s.registrations.Load(id); ok {
		reg := val.(models.Registration)
		reg.Status = status
		s.registrations.Store(id, reg)
		return true
	}
	return false
}

func (s *Storage) SavePrescription(pr *models.Prescription) {
	s.prescriptions.Store(pr.ID, *pr)
}

func (s *Storage) GetPrescription(id string) (models.Prescription, bool) {
	if val, ok := s.prescriptions.Load(id); ok {
		return val.(models.Prescription), true
	}
	return models.Prescription{}, false
}

func (s *Storage) GetPrescriptionsByRegistration(regID string) []models.Prescription {
	var list []models.Prescription
	s.prescriptions.Range(func(key, value interface{}) bool {
		p := value.(models.Prescription)
		if p.RegistrationID == regID {
			list = append(list, p)
		}
		return true
	})
	return list
}

func (s *Storage) GetAllPrescriptions() []models.Prescription {
	var list []models.Prescription
	s.prescriptions.Range(func(key, value interface{}) bool {
		p := value.(models.Prescription)
		list = append(list, p)
		return true
	})
	return list
}

func (s *Storage) GetHerb(name string) (models.Herb, bool) {
	if val, ok := s.herbs.Load(name); ok {
		return val.(models.Herb), true
	}
	return models.Herb{}, false
}

func (s *Storage) SaveHerb(herb models.Herb) {
	s.herbs.Store(herb.Name, herb)
}

func (s *Storage) GetAllHerbs() []models.Herb {
	var herbs []models.Herb
	s.herbs.Range(func(key, value interface{}) bool {
		herbs = append(herbs, value.(models.Herb))
		return true
	})
	return herbs
}

func (s *Storage) DeductStock(items []models.PrescriptionItem) (bool, map[string]float64) {
	missing := make(map[string]float64)

	for _, item := range items {
		herb, ok := s.GetHerb(item.HerbName)
		if !ok {
			missing[item.HerbName] = 0
			continue
		}
		if herb.Stock < item.Dosage {
			missing[item.HerbName] = herb.Stock
		}
	}

	if len(missing) > 0 {
		return false, missing
	}

	for _, item := range items {
		herb, _ := s.GetHerb(item.HerbName)
		herb.Stock -= item.Dosage
		s.SaveHerb(herb)
	}

	return true, nil
}
