package service

import (
	"sort"
	"time"

	"credits/pkg/models"
	"credits/pkg/storage"
)

const (
	RequiredCreditsRequired = 20
	ElectiveCreditsRequired = 10
	TotalCreditsRequired    = RequiredCreditsRequired + ElectiveCreditsRequired
	MaxCarryOverCredits     = 5
)

type CreditService struct {
	store *storage.Storage
}

func NewCreditService(store *storage.Storage) *CreditService {
	return &CreditService{store: store}
}

func (s *CreditService) CalculateProgress(employeeID string, year int) (*models.CreditProgress, error) {
	employee, err := s.store.GetEmployee(employeeID)
	if err != nil {
		return nil, err
	}

	records := s.store.GetCreditRecordsByEmployeeAndYear(employeeID, year)
	courses := s.store.ListCourses()

	courseMap := make(map[string]*models.Course)
	for _, c := range courses {
		courseMap[c.ID] = c
	}

	requiredCredits := 0
	electiveCredits := 0
	validRecords := 0

	for _, record := range records {
		course, exists := courseMap[record.CourseID]
		if exists && !course.ValidUntil.Before(record.CompletedAt) {
			validRecords++
			if course.Type == models.RequiredCourse {
				requiredCredits += course.Credits
			} else {
				electiveCredits += course.Credits
			}
		}
	}

	carryOverAvailable := 0
	if employee.CarryOverCredits != nil {
		if co, exists := employee.CarryOverCredits[year]; exists {
			carryOverAvailable = co
		}
	}

	carryOverUsed := 0
	totalCredits := requiredCredits + electiveCredits
	if totalCredits < ElectiveCreditsRequired && carryOverAvailable > 0 {
		needed := ElectiveCreditsRequired - electiveCredits
		if carryOverAvailable >= needed {
			carryOverUsed = needed
		} else {
			carryOverUsed = carryOverAvailable
		}
	}

	allRequiredCoursesDone := requiredCredits >= RequiredCreditsRequired

	progressPercentage := 0.0
	if TotalCreditsRequired > 0 {
		progressPercentage = float64(totalCredits+carryOverUsed) / float64(TotalCreditsRequired) * 100
		if progressPercentage > 100 {
			progressPercentage = 100
		}
	}

	return &models.CreditProgress{
		EmployeeID:               employeeID,
		EmployeeName:             employee.Name,
		Department:               employee.Department,
		Year:                     year,
		RequiredCreditsCompleted: requiredCredits,
		RequiredCreditsRequired:  RequiredCreditsRequired,
		ElectiveCreditsCompleted: electiveCredits,
		ElectiveCreditsRequired:  ElectiveCreditsRequired,
		CarryOverUsed:            carryOverUsed,
		CarryOverAvailable:       carryOverAvailable,
		TotalCreditsCompleted:    totalCredits + carryOverUsed,
		TotalCreditsRequired:     TotalCreditsRequired,
		AllRequiredCoursesDone:   allRequiredCoursesDone,
		ProgressPercentage:       progressPercentage,
	}, nil
}

func (s *CreditService) CalculateCarryOver(employeeID string, year int) (int, error) {
	records := s.store.GetCreditRecordsByEmployeeAndYear(employeeID, year)
	courses := s.store.ListCourses()

	courseMap := make(map[string]*models.Course)
	for _, c := range courses {
		courseMap[c.ID] = c
	}

	requiredCredits := 0
	electiveCredits := 0

	for _, record := range records {
		course, exists := courseMap[record.CourseID]
		if exists && !course.ValidUntil.Before(record.CompletedAt) {
			if course.Type == models.RequiredCourse {
				requiredCredits += course.Credits
			} else {
				electiveCredits += course.Credits
			}
		}
	}

	excessElective := 0
	if requiredCredits >= RequiredCreditsRequired {
		if electiveCredits > ElectiveCreditsRequired {
			excessElective = electiveCredits - ElectiveCreditsRequired
		}
	}

	carryOver := excessElective
	if carryOver > MaxCarryOverCredits {
		carryOver = MaxCarryOverCredits
	}

	return carryOver, nil
}

func (s *CreditService) GetDepartmentRanking(department string, year int) ([]*models.RankItem, error) {
	employees := s.store.GetEmployeesByDepartment(department)

	rankItems := make([]*models.RankItem, 0, len(employees))

	for _, emp := range employees {
		progress, err := s.CalculateProgress(emp.ID, year)
		if err != nil {
			continue
		}

		rankItems = append(rankItems, &models.RankItem{
			EmployeeID:             emp.ID,
			EmployeeName:           emp.Name,
			Department:             emp.Department,
			TotalCredits:           progress.TotalCreditsCompleted,
			RequiredCredits:        progress.RequiredCreditsCompleted,
			ElectiveCredits:        progress.ElectiveCreditsCompleted,
			AllRequiredCoursesDone: progress.AllRequiredCoursesDone,
			ProgressPercentage:     progress.ProgressPercentage,
		})
	}

	sort.Slice(rankItems, func(i, j int) bool {
		if rankItems[i].TotalCredits != rankItems[j].TotalCredits {
			return rankItems[i].TotalCredits > rankItems[j].TotalCredits
		}
		if rankItems[i].AllRequiredCoursesDone != rankItems[j].AllRequiredCoursesDone {
			return rankItems[i].AllRequiredCoursesDone
		}
		return rankItems[i].ProgressPercentage > rankItems[j].ProgressPercentage
	})

	for i := range rankItems {
		rankItems[i].Rank = i + 1
	}

	if len(rankItems) == 0 {
		return []*models.RankItem{}, nil
	}

	return rankItems, nil
}

func (s *CreditService) GetYearlySummary(employeeID string, year int) (*models.YearlySummary, error) {
	progress, err := s.CalculateProgress(employeeID, year)
	if err != nil {
		return nil, err
	}

	carryOver, err := s.CalculateCarryOver(employeeID, year)
	if err != nil {
		return nil, err
	}

	return &models.YearlySummary{
		Year:                     year,
		RequiredCreditsCompleted: progress.RequiredCreditsCompleted,
		ElectiveCreditsCompleted: progress.ElectiveCreditsCompleted,
		CarryOverUsed:            progress.CarryOverUsed,
		TotalCreditsCompleted:    progress.TotalCreditsCompleted,
		RequiredCreditsRequired:  progress.RequiredCreditsRequired,
		ElectiveCreditsRequired:  progress.ElectiveCreditsRequired,
		TotalCreditsRequired:     progress.TotalCreditsRequired,
		AllRequiredCoursesDone:   progress.AllRequiredCoursesDone,
		CarryOverToNextYear:      carryOver,
	}, nil
}

func (s *CreditService) ProcessYearEndCarryOver(employeeID string, year int) error {
	carryOver, err := s.CalculateCarryOver(employeeID, year)
	if err != nil {
		return err
	}

	employee, err := s.store.GetEmployee(employeeID)
	if err != nil {
		return err
	}

	if employee.CarryOverCredits == nil {
		employee.CarryOverCredits = make(map[int]int)
	}
	employee.CarryOverCredits[year+1] = carryOver
	employee.UpdatedAt = time.Now()

	return s.store.UpdateEmployee(employee)
}
