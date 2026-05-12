package repository

import (
	"fmt"
	"medical-exam-system/internal/model"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Repository struct {
	mu              sync.RWMutex
	items           map[string]*model.ExamItem
	packages        map[string]*model.Package
	appointments    map[string]*model.Appointment
	examResults     map[string]*model.ExamResult
	criticalAlerts  []*model.CriticalAlert
	reports         map[string]*model.Report
	examNumberCounter map[string]uint64
	timeSlotCount   map[string]map[model.TimeSlot]int
}

func New() *Repository {
	repo := &Repository{
		items:             make(map[string]*model.ExamItem),
		packages:          make(map[string]*model.Package),
		appointments:      make(map[string]*model.Appointment),
		examResults:       make(map[string]*model.ExamResult),
		criticalAlerts:    []*model.CriticalAlert{},
		reports:           make(map[string]*model.Report),
		examNumberCounter: make(map[string]uint64),
		timeSlotCount:     make(map[string]map[model.TimeSlot]int),
	}
	repo.initDefaultData()
	return repo
}

func (r *Repository) initDefaultData() {
	items := []*model.ExamItem{
		{ID: uuid.New().String(), Name: "血常规", Department: model.DeptLab, Unit: "", Price: 50},
		{ID: uuid.New().String(), Name: "肝功能", Department: model.DeptLab, Unit: "", Price: 80},
		{ID: uuid.New().String(), Name: "肾功能", Department: model.DeptLab, Unit: "", Price: 60},
		{ID: uuid.New().String(), Name: "血脂", Department: model.DeptLab, Unit: "", Price: 50},
		{ID: uuid.New().String(), Name: "血糖", Department: model.DeptLab, Unit: "mmol/L", RefRange: model.RefRange{Min: floatPtr(3.9), Max: floatPtr(6.1)}, Price: 30},
		{ID: uuid.New().String(), Name: "肿瘤标志物", Department: model.DeptLab, Unit: "", RefRange: model.RefRange{Max: floatPtr(5.0)}, Price: 150},
		{ID: uuid.New().String(), Name: "血红蛋白", Department: model.DeptLab, Unit: "g/L", RefRange: model.RefRange{Min: floatPtr(110)}, Price: 40},
		{ID: uuid.New().String(), Name: "胸部CT", Department: model.DeptImaging, Unit: "", Price: 300},
		{ID: uuid.New().String(), Name: "腹部B超", Department: model.DeptUltrasound, Unit: "", Price: 150},
		{ID: uuid.New().String(), Name: "心电图", Department: model.DeptECG, Unit: "", Price: 40},
		{ID: uuid.New().String(), Name: "视力检查", Department: model.DeptOphthalmology, Unit: "", Price: 30},
		{ID: uuid.New().String(), Name: "耳鼻喉检查", Department: model.DeptENT, Unit: "", Price: 40},
	}
	for _, item := range items {
		r.items[item.ID] = item
	}

	itemIDs := make([]string, len(items))
	for i, item := range items {
		itemIDs[i] = item.ID
	}

	basicPkg := &model.Package{
		ID:          uuid.New().String(),
		Name:        "基础套餐",
		Price:       59900,
		Description: "适合年轻人群的基础体检",
		ItemIDs:     []string{itemIDs[0], itemIDs[1], itemIDs[4], itemIDs[8], itemIDs[9], itemIDs[10]},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	r.packages[basicPkg.ID] = basicPkg

	middlePkg := &model.Package{
		ID:          uuid.New().String(),
		Name:        "中老年套餐",
		Price:       129900,
		Description: "针对中老年人群的全面体检",
		ItemIDs:     itemIDs[:8],
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	r.packages[middlePkg.ID] = middlePkg

	womenPkg := &model.Package{
		ID:          uuid.New().String(),
		Name:        "女性专属套餐",
		Price:       89900,
		Description: "专为女性设计的体检套餐",
		ItemIDs:     []string{itemIDs[0], itemIDs[1], itemIDs[2], itemIDs[4], itemIDs[7], itemIDs[8], itemIDs[9]},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	r.packages[womenPkg.ID] = womenPkg

	deepPkg := &model.Package{
		ID:          uuid.New().String(),
		Name:        "深度体检套餐",
		Price:       299900,
		Description: "高端深度体检，全面筛查",
		ItemIDs:     itemIDs,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	r.packages[deepPkg.ID] = deepPkg
}

func floatPtr(f float64) *float64 {
	return &f
}

func (r *Repository) GetItemByID(id string) (*model.ExamItem, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.items[id]
	return item, ok
}

func (r *Repository) GetAllItems() []*model.ExamItem {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]*model.ExamItem, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, item)
	}
	return items
}

func (r *Repository) GetPackages(page, size int) ([]*model.Package, int64) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	pkgs := make([]*model.Package, 0, len(r.packages))
	for _, pkg := range r.packages {
		pkgs = append(pkgs, pkg)
	}
	total := int64(len(pkgs))
	start := (page - 1) * size
	if start < 0 {
		start = 0
	}
	end := start + size
	if start >= len(pkgs) {
		return []*model.Package{}, total
	}
	if end > len(pkgs) {
		end = len(pkgs)
	}
	return pkgs[start:end], total
}

func (r *Repository) GetPackageByID(id string) (*model.Package, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	pkg, ok := r.packages[id]
	return pkg, ok
}

func (r *Repository) CreatePackage(pkg *model.Package) {
	r.mu.Lock()
	defer r.mu.Unlock()
	pkg.ID = uuid.New().String()
	pkg.CreatedAt = time.Now()
	pkg.UpdatedAt = time.Now()
	r.packages[pkg.ID] = pkg
}

func (r *Repository) UpdatePackage(pkg *model.Package) {
	r.mu.Lock()
	defer r.mu.Unlock()
	pkg.UpdatedAt = time.Now()
	r.packages[pkg.ID] = pkg
}

func (r *Repository) DeletePackage(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.packages, id)
}

func (r *Repository) GetTimeSlotCount(date string, slot model.TimeSlot) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if slots, ok := r.timeSlotCount[date]; ok {
		return slots[slot]
	}
	return 0
}

func (r *Repository) IncrementTimeSlotCount(date string, slot model.TimeSlot) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	max := 90
	if slot == model.TimeSlotAfternoon {
		max = 60
	}
	if _, ok := r.timeSlotCount[date]; !ok {
		r.timeSlotCount[date] = make(map[model.TimeSlot]int)
	}
	current := r.timeSlotCount[date][slot]
	if current >= max {
		return false
	}
	r.timeSlotCount[date][slot] = current + 1
	return true
}

func (r *Repository) DecrementTimeSlotCount(date string, slot model.TimeSlot) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if slots, ok := r.timeSlotCount[date]; ok {
		if slots[slot] > 0 {
			slots[slot]--
		}
	}
}

func (r *Repository) GenerateExamNumber(date string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	dateKey := date
	counter := r.examNumberCounter[dateKey]
	counter++
	r.examNumberCounter[dateKey] = counter
	return "TJ" + date + fmt.Sprintf("%04d", counter)
}

func (r *Repository) CreateAppointment(appt *model.Appointment) {
	r.mu.Lock()
	defer r.mu.Unlock()
	appt.ID = uuid.New().String()
	appt.CreatedAt = time.Now()
	appt.UpdatedAt = time.Now()
	r.appointments[appt.ID] = appt
}

func (r *Repository) GetAppointmentByID(id string) (*model.Appointment, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	appt, ok := r.appointments[id]
	return appt, ok
}

func (r *Repository) GetAppointments(page, size int) ([]*model.Appointment, int64) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	appts := make([]*model.Appointment, 0, len(r.appointments))
	for _, appt := range r.appointments {
		appts = append(appts, appt)
	}
	total := int64(len(appts))
	start := (page - 1) * size
	if start < 0 {
		start = 0
	}
	end := start + size
	if start >= len(appts) {
		return []*model.Appointment{}, total
	}
	if end > len(appts) {
		end = len(appts)
	}
	return appts[start:end], total
}

func (r *Repository) UpdateAppointment(appt *model.Appointment) {
	r.mu.Lock()
	defer r.mu.Unlock()
	appt.UpdatedAt = time.Now()
	r.appointments[appt.ID] = appt
}

func (r *Repository) CreateExamResult(result *model.ExamResult) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result.ID = uuid.New().String()
	result.CreatedAt = time.Now()
	result.UpdatedAt = time.Now()
	r.examResults[result.ID] = result
}

func (r *Repository) GetExamResultByID(id string) (*model.ExamResult, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result, ok := r.examResults[id]
	return result, ok
}

func (r *Repository) GetExamResultsByAppointment(appointmentID string) []*model.ExamResult {
	r.mu.RLock()
	defer r.mu.RUnlock()
	results := []*model.ExamResult{}
	for _, result := range r.examResults {
		if result.AppointmentID == appointmentID {
			results = append(results, result)
		}
	}
	return results
}

func (r *Repository) UpdateExamResult(result *model.ExamResult) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result.UpdatedAt = time.Now()
	r.examResults[result.ID] = result
}

func (r *Repository) CreateCriticalAlert(alert *model.CriticalAlert) {
	r.mu.Lock()
	defer r.mu.Unlock()
	alert.ID = uuid.New().String()
	alert.CreatedAt = time.Now()
	r.criticalAlerts = append(r.criticalAlerts, alert)
}

func (r *Repository) GetUnresolvedCriticalAlerts() []*model.CriticalAlert {
	r.mu.RLock()
	defer r.mu.RUnlock()
	alerts := []*model.CriticalAlert{}
	for _, alert := range r.criticalAlerts {
		if !alert.Resolved {
			alerts = append(alerts, alert)
		}
	}
	return alerts
}

func (r *Repository) GetRecentCriticalAlerts(limit int) []*model.CriticalAlert {
	r.mu.RLock()
	defer r.mu.RUnlock()
	start := len(r.criticalAlerts) - limit
	if start < 0 {
		start = 0
	}
	return r.criticalAlerts[start:]
}

func (r *Repository) CreateReport(report *model.Report) {
	r.mu.Lock()
	defer r.mu.Unlock()
	report.ID = uuid.New().String()
	report.CreatedAt = time.Now()
	report.UpdatedAt = time.Now()
	r.reports[report.ID] = report
}

func (r *Repository) GetReportByID(id string) (*model.Report, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	report, ok := r.reports[id]
	return report, ok
}

func (r *Repository) GetReportByAppointment(appointmentID string) (*model.Report, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, report := range r.reports {
		if report.AppointmentID == appointmentID {
			return report, true
		}
	}
	return nil, false
}

func (r *Repository) GetReports(page, size int) ([]*model.Report, int64) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	reports := make([]*model.Report, 0, len(r.reports))
	for _, report := range r.reports {
		reports = append(reports, report)
	}
	total := int64(len(reports))
	start := (page - 1) * size
	if start < 0 {
		start = 0
	}
	end := start + size
	if start >= len(reports) {
		return []*model.Report{}, total
	}
	if end > len(reports) {
		end = len(reports)
	}
	return reports[start:end], total
}

func (r *Repository) UpdateReport(report *model.Report) {
	r.mu.Lock()
	defer r.mu.Unlock()
	report.UpdatedAt = time.Now()
	r.reports[report.ID] = report
}

func (r *Repository) GetReportsByDateRange(startDate, endDate time.Time) []*model.Report {
	r.mu.RLock()
	defer r.mu.RUnlock()
	reports := []*model.Report{}
	for _, report := range r.reports {
		if report.CreatedAt.After(startDate) && report.CreatedAt.Before(endDate.AddDate(0, 0, 1)) {
			reports = append(reports, report)
		}
	}
	return reports
}

func (r *Repository) GetAllAppointments() []*model.Appointment {
	r.mu.RLock()
	defer r.mu.RUnlock()
	appts := make([]*model.Appointment, 0, len(r.appointments))
	for _, appt := range r.appointments {
		appts = append(appts, appt)
	}
	return appts
}
