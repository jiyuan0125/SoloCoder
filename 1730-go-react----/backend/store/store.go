package store

import (
	"encoding/json"
	"sync"
	"time"

	"diploma-auth-system/models"

	"github.com/google/uuid"
)

type Store struct {
	mu       sync.RWMutex
	users    map[string]models.User
	diplomas map[string]models.Diploma
	logs     map[string]models.OperationLog
}

func NewStore() *Store {
	s := &Store{
		users:    make(map[string]models.User),
		diplomas: make(map[string]models.Diploma),
		logs:     make(map[string]models.OperationLog),
	}
	
	adminUser := models.User{
		ID:       uuid.New().String(),
		Username: "admin",
		Password: "admin123",
		Role:     models.RoleAdmin,
	}
	s.users[adminUser.ID] = adminUser
	
	verifierUser := models.User{
		ID:       uuid.New().String(),
		Username: "verifier",
		Password: "verifier123",
		Role:     models.RoleVerifier,
	}
	s.users[verifierUser.ID] = verifierUser
	
	viewerUser := models.User{
		ID:       uuid.New().String(),
		Username: "viewer",
		Password: "viewer123",
		Role:     models.RoleViewer,
	}
	s.users[viewerUser.ID] = viewerUser
	
	return s
}

func (s *Store) Authenticate(username, password string) (*models.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	for _, u := range s.users {
		if u.Username == username && u.Password == password {
			user := u
			user.Password = ""
			return &user, true
		}
	}
	return nil, false
}

func (s *Store) GetUser(id string) (*models.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	user, exists := s.users[id]
	if !exists {
		return nil, false
	}
	user.Password = ""
	return &user, true
}

func (s *Store) ListUsers() []models.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	users := make([]models.User, 0, len(s.users))
	for _, u := range s.users {
		user := u
		user.Password = ""
		users = append(users, user)
	}
	return users
}

func (s *Store) CreateUser(req models.CreateUserRequest) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	for _, u := range s.users {
		if u.Username == req.Username {
			return nil, &DuplicateError{Message: "用户名已存在"}
		}
	}
	
	user := models.User{
		ID:       uuid.New().String(),
		Username: req.Username,
		Password: req.Password,
		Role:     req.Role,
	}
	s.users[user.ID] = user
	
	result := user
	result.Password = ""
	return &result, nil
}

func (s *Store) UpdateUser(id string, req models.UpdateUserRequest) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	user, exists := s.users[id]
	if !exists {
		return nil, &NotFoundError{Message: "用户不存在"}
	}
	
	if req.Password != nil {
		user.Password = *req.Password
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	
	s.users[id] = user
	
	result := user
	result.Password = ""
	return &result, nil
}

func (s *Store) DeleteUser(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, exists := s.users[id]; !exists {
		return &NotFoundError{Message: "用户不存在"}
	}
	
	delete(s.users, id)
	return nil
}

func (s *Store) ListDiplomas() []models.Diploma {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	diplomas := make([]models.Diploma, 0, len(s.diplomas))
	for _, d := range s.diplomas {
		diplomas = append(diplomas, d)
	}
	return diplomas
}

func (s *Store) GetDiploma(id string) (*models.Diploma, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	diploma, exists := s.diplomas[id]
	if !exists {
		return nil, false
	}
	return &diploma, true
}

func (s *Store) CheckDuplicate(name, idCard string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	trimmedIDCard := idCard
	for _, d := range s.diplomas {
		if d.Name == name && d.IDCard == trimmedIDCard {
			return true
		}
	}
	return false
}

func (s *Store) CheckDuplicateExclude(id, name, idCard string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	trimmedIDCard := idCard
	for _, d := range s.diplomas {
		if d.ID != id && d.Name == name && d.IDCard == trimmedIDCard {
			return true
		}
	}
	return false
}

func (s *Store) CreateDiploma(req models.CreateDiplomaRequest) (*models.Diploma, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	now := time.Now()
	diploma := models.Diploma{
		ID:              uuid.New().String(),
		Name:            req.Name,
		IDCard:          req.IDCard,
		School:          req.School,
		Level:           req.Level,
		Major:           req.Major,
		StudyYears:      req.StudyYears,
		EnrollmentDate:  req.EnrollmentDate,
		GraduationDate:  req.GraduationDate,
		Status:          models.StatusValid,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	
	s.diplomas[diploma.ID] = diploma
	return &diploma, nil
}

func (s *Store) UpdateDiploma(id string, req models.UpdateDiplomaRequest) (*models.Diploma, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	diploma, exists := s.diplomas[id]
	if !exists {
		return nil, &NotFoundError{Message: "学历信息不存在"}
	}
	
	if req.Name != nil {
		diploma.Name = *req.Name
	}
	if req.IDCard != nil {
		diploma.IDCard = *req.IDCard
	}
	if req.School != nil {
		diploma.School = *req.School
	}
	if req.Level != nil {
		diploma.Level = *req.Level
	}
	if req.Major != nil {
		diploma.Major = *req.Major
	}
	if req.StudyYears != nil {
		diploma.StudyYears = *req.StudyYears
	}
	if req.EnrollmentDate != nil {
		diploma.EnrollmentDate = *req.EnrollmentDate
	}
	if req.GraduationDate != nil {
		diploma.GraduationDate = *req.GraduationDate
	}
	if req.Status != nil {
		diploma.Status = *req.Status
	}
	
	diploma.UpdatedAt = time.Now()
	s.diplomas[id] = diploma
	
	return &diploma, nil
}

func (s *Store) DeleteDiploma(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, exists := s.diplomas[id]; !exists {
		return &NotFoundError{Message: "学历信息不存在"}
	}
	
	delete(s.diplomas, id)
	return nil
}

func (s *Store) VerifyDiploma(name, idCard string) []models.Diploma {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var results []models.Diploma
	for _, d := range s.diplomas {
		if d.Name == name && d.IDCard == idCard {
			results = append(results, d)
		}
	}
	return results
}

func (s *Store) AddLog(operator *models.User, opType models.OperationType, content string, result *models.VerificationResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	log := models.OperationLog{
		ID:                 uuid.New().String(),
		OperatorID:         operator.ID,
		OperatorName:       operator.Username,
		OperatorRole:       operator.Role,
		OperationType:      opType,
		Content:            content,
		VerificationResult: result,
		CreatedAt:          time.Now(),
	}
	s.logs[log.ID] = log
}

func (s *Store) ListLogs(filter models.LogFilter) (*models.PagedResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if filter.Page < 1 {
		return nil, &ValidationError{Message: "页码必须大于等于1"}
	}
	if filter.PageSize != 20 {
		return nil, &ValidationError{Message: "每页条数必须为20"}
	}
	
	var filteredLogs []models.OperationLog
	for _, log := range s.logs {
		match := true
		
		if filter.OperationType != nil && log.OperationType != *filter.OperationType {
			match = false
		}
		
		if filter.StartDate != nil && log.CreatedAt.Before(*filter.StartDate) {
			match = false
		}
		
		if filter.EndDate != nil && log.CreatedAt.After(*filter.EndDate) {
			match = false
		}
		
		if match {
			filteredLogs = append(filteredLogs, log)
		}
	}
	
	for i := 0; i < len(filteredLogs)-1; i++ {
		for j := i + 1; j < len(filteredLogs); j++ {
			if filteredLogs[i].CreatedAt.Before(filteredLogs[j].CreatedAt) {
				filteredLogs[i], filteredLogs[j] = filteredLogs[j], filteredLogs[i]
			}
		}
	}
	
	total := len(filteredLogs)
	start := (filter.Page - 1) * filter.PageSize
	end := start + filter.PageSize
	
	if start >= total {
		return &models.PagedResponse{
			Data:     []models.OperationLog{},
			Total:    total,
			Page:     filter.Page,
			PageSize: filter.PageSize,
		}, nil
	}
	
	if end > total {
		end = total
	}
	
	return &models.PagedResponse{
		Data:     filteredLogs[start:end],
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func (s *Store) LogDiplomaChanges(old, new *models.Diploma) string {
	type change struct {
		Field    string      `json:"field"`
		OldValue interface{} `json:"old_value"`
		NewValue interface{} `json:"new_value"`
	}
	
	var changes []change
	
	if old.Name != new.Name {
		changes = append(changes, change{Field: "name", OldValue: old.Name, NewValue: new.Name})
	}
	if old.IDCard != new.IDCard {
		changes = append(changes, change{Field: "id_card", OldValue: old.IDCard, NewValue: new.IDCard})
	}
	if old.School != new.School {
		changes = append(changes, change{Field: "school", OldValue: old.School, NewValue: new.School})
	}
	if old.Level != new.Level {
		changes = append(changes, change{Field: "level", OldValue: old.Level, NewValue: new.Level})
	}
	if old.Major != new.Major {
		changes = append(changes, change{Field: "major", OldValue: old.Major, NewValue: new.Major})
	}
	if old.StudyYears != new.StudyYears {
		changes = append(changes, change{Field: "study_years", OldValue: old.StudyYears, NewValue: new.StudyYears})
	}
	if old.EnrollmentDate != new.EnrollmentDate {
		changes = append(changes, change{Field: "enrollment_date", OldValue: old.EnrollmentDate, NewValue: new.EnrollmentDate})
	}
	if old.GraduationDate != new.GraduationDate {
		changes = append(changes, change{Field: "graduation_date", OldValue: old.GraduationDate, NewValue: new.GraduationDate})
	}
	if old.Status != new.Status {
		changes = append(changes, change{Field: "status", OldValue: old.Status, NewValue: new.Status})
	}
	
	data, _ := json.Marshal(changes)
	return string(data)
}

type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

type DuplicateError struct {
	Message string
}

func (e *DuplicateError) Error() string {
	return e.Message
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
