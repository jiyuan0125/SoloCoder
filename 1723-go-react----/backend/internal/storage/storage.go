package storage

import (
	"sync"

	"learning-platform/internal/models"
	"learning-platform/pkg/dag"
)

type Storage struct {
	mu               sync.RWMutex
	courses          map[string]*models.Course
	students         map[string]*models.Student
	enrollments      map[string]*models.Enrollment
	learningProgress map[string]*models.LearningProgress
	quizAttempts     map[string]*models.QuizAttempt
	learningRecords  map[string]*models.LearningRecord
	achievements     map[string]*models.Achievement
	products         map[string]*models.Product
	priceHistory     map[string]*models.PriceHistory
	orders           map[string]*models.Order
	purchaseRequests map[string]*models.PurchaseRequest
	courseGraph      *dag.Graph
}

func NewStorage() *Storage {
	return &Storage{
		courses:          make(map[string]*models.Course),
		students:         make(map[string]*models.Student),
		enrollments:      make(map[string]*models.Enrollment),
		learningProgress: make(map[string]*models.LearningProgress),
		quizAttempts:     make(map[string]*models.QuizAttempt),
		learningRecords:  make(map[string]*models.LearningRecord),
		achievements:     make(map[string]*models.Achievement),
		products:         make(map[string]*models.Product),
		priceHistory:     make(map[string]*models.PriceHistory),
		orders:           make(map[string]*models.Order),
		purchaseRequests: make(map[string]*models.PurchaseRequest),
		courseGraph:      dag.NewGraph(),
	}
}

func (s *Storage) GetCourseGraph() *dag.Graph {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.courseGraph
}

func (s *Storage) CreateCourse(course *models.Course) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.courses[course.ID]; exists {
		return &DuplicateError{ID: course.ID, Type: "course"}
	}

	s.courses[course.ID] = course
	return s.courseGraph.AddNode(course.ID, course)
}

func (s *Storage) GetCourse(id string) (*models.Course, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	course, exists := s.courses[id]
	return course, exists
}

func (s *Storage) GetAllCourses() map[string]*models.Course {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.courses
}

func (s *Storage) AddPrerequisite(courseID, prereqID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.courses[courseID]; !exists {
		return &NotFoundError{ID: courseID, Type: "course"}
	}
	if _, exists := s.courses[prereqID]; !exists {
		return &NotFoundError{ID: prereqID, Type: "course"}
	}

	if err := s.courseGraph.AddEdge(prereqID, courseID); err != nil {
		return &CycleError{Message: err.Error()}
	}

	course := s.courses[courseID]
	for _, id := range course.PrerequisiteIDs {
		if id == prereqID {
			return nil
		}
	}
	course.PrerequisiteIDs = append(course.PrerequisiteIDs, prereqID)

	return nil
}

func (s *Storage) CreateStudent(student *models.Student) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.students[student.ID]; exists {
		return &DuplicateError{ID: student.ID, Type: "student"}
	}

	s.students[student.ID] = student
	return nil
}

func (s *Storage) GetStudent(id string) (*models.Student, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	student, exists := s.students[id]
	return student, exists
}

func (s *Storage) CreateEnrollment(enrollment *models.Enrollment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.enrollments[enrollment.ID]; exists {
		return &DuplicateError{ID: enrollment.ID, Type: "enrollment"}
	}

	s.enrollments[enrollment.ID] = enrollment
	return nil
}

func (s *Storage) GetEnrollmentsByStudent(studentID string) []*models.Enrollment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*models.Enrollment
	for _, e := range s.enrollments {
		if e.StudentID == studentID {
			result = append(result, e)
		}
	}
	return result
}

func (s *Storage) CreateLearningProgress(progress *models.LearningProgress) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.learningProgress[progress.ID]; exists {
		return &DuplicateError{ID: progress.ID, Type: "learning_progress"}
	}

	s.learningProgress[progress.ID] = progress
	return nil
}

func (s *Storage) GetLearningProgress(studentID, courseID string) (*models.LearningProgress, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, p := range s.learningProgress {
		if p.StudentID == studentID && p.CourseID == courseID {
			return p, true
		}
	}
	return nil, false
}

func (s *Storage) UpdateLearningProgress(progress *models.LearningProgress) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.learningProgress[progress.ID] = progress
}

func (s *Storage) CreateQuizAttempt(attempt *models.QuizAttempt) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.quizAttempts[attempt.ID]; exists {
		return &DuplicateError{ID: attempt.ID, Type: "quiz_attempt"}
	}

	s.quizAttempts[attempt.ID] = attempt
	return nil
}

func (s *Storage) GetQuizAttempts(studentID, courseID string) []*models.QuizAttempt {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*models.QuizAttempt
	for _, a := range s.quizAttempts {
		if a.StudentID == studentID && a.CourseID == courseID {
			result = append(result, a)
		}
	}
	return result
}

func (s *Storage) CreateLearningRecord(record *models.LearningRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.learningRecords[record.ID]; exists {
		return &DuplicateError{ID: record.ID, Type: "learning_record"}
	}

	s.learningRecords[record.ID] = record
	return nil
}

func (s *Storage) GetLearningRecordsByStudent(studentID string) []*models.LearningRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*models.LearningRecord
	for _, r := range s.learningRecords {
		if r.StudentID == studentID {
			result = append(result, r)
		}
	}
	return result
}

func (s *Storage) CreateAchievement(achievement *models.Achievement) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.achievements[achievement.ID]; exists {
		return &DuplicateError{ID: achievement.ID, Type: "achievement"}
	}

	s.achievements[achievement.ID] = achievement
	return nil
}

func (s *Storage) GetAchievementsByStudent(studentID string) []*models.Achievement {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*models.Achievement
	for _, a := range s.achievements {
		if a.StudentID == studentID {
			result = append(result, a)
		}
	}
	return result
}

func (s *Storage) HasAchievement(studentID, achievementType string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, a := range s.achievements {
		if a.StudentID == studentID && a.Type == achievementType {
			return true
		}
	}
	return false
}

func (s *Storage) CreateProduct(product *models.Product) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.products[product.ID]; exists {
		return &DuplicateError{ID: product.ID, Type: "product"}
	}

	s.products[product.ID] = product
	return nil
}

func (s *Storage) GetProduct(id string) (*models.Product, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	product, exists := s.products[id]
	return product, exists
}

func (s *Storage) GetAllProducts() []*models.Product {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*models.Product
	for _, p := range s.products {
		result = append(result, p)
	}
	return result
}

func (s *Storage) UpdateProduct(product *models.Product) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.products[product.ID] = product
}

func (s *Storage) CreatePriceHistory(history *models.PriceHistory) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.priceHistory[history.ID]; exists {
		return &DuplicateError{ID: history.ID, Type: "price_history"}
	}

	s.priceHistory[history.ID] = history
	return nil
}

func (s *Storage) GetPriceHistoryByProduct(productID string) []*models.PriceHistory {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*models.PriceHistory
	for _, h := range s.priceHistory {
		if h.ProductID == productID {
			result = append(result, h)
		}
	}
	return result
}

func (s *Storage) CreateOrder(order *models.Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.orders[order.ID]; exists {
		return &DuplicateError{ID: order.ID, Type: "order"}
	}

	s.orders[order.ID] = order
	return nil
}

func (s *Storage) GetOrder(id string) (*models.Order, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	order, exists := s.orders[id]
	return order, exists
}

func (s *Storage) GetAllOrders() []*models.Order {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*models.Order
	for _, o := range s.orders {
		result = append(result, o)
	}
	return result
}

func (s *Storage) UpdateOrder(order *models.Order) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders[order.ID] = order
}

func (s *Storage) CreatePurchaseRequest(request *models.PurchaseRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.purchaseRequests[request.ID]; exists {
		return &DuplicateError{ID: request.ID, Type: "purchase_request"}
	}

	s.purchaseRequests[request.ID] = request
	return nil
}

func (s *Storage) GetAllPurchaseRequests() []*models.PurchaseRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*models.PurchaseRequest
	for _, r := range s.purchaseRequests {
		result = append(result, r)
	}
	return result
}

type DuplicateError struct {
	ID   string
	Type string
}

func (e *DuplicateError) Error() string {
	return e.Type + " with id " + e.ID + " already exists"
}

type NotFoundError struct {
	ID   string
	Type string
}

func (e *NotFoundError) Error() string {
	return e.Type + " with id " + e.ID + " not found"
}

type CycleError struct {
	Message string
}

func (e *CycleError) Error() string {
	return e.Message
}
