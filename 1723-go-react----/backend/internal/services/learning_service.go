package services

import (
	"learning-platform/internal/models"
	"learning-platform/internal/storage"
	"learning-platform/internal/utils"
	"learning-platform/pkg/dag"
	"math/rand"
	"sort"
	"time"
)

const MaxConcurrentCourses = 5

type LearningService struct {
	storage        *storage.Storage
	courseService  *CourseService
	achievementSvc *AchievementService
}

func NewLearningService(s *storage.Storage, cs *CourseService, as *AchievementService) *LearningService {
	return &LearningService{
		storage:        s,
		courseService:  cs,
		achievementSvc: as,
	}
}

func (s *LearningService) CreateStudent(student *models.Student) (*models.Student, error) {
	if student.ID == "" {
		student.ID = utils.GenerateUUID()
	}
	student.CreatedAt = time.Now()

	err := s.storage.CreateStudent(student)
	if err != nil {
		return nil, err
	}
	return student, nil
}

func (s *LearningService) GetStudent(id string) (*models.Student, bool) {
	return s.storage.GetStudent(id)
}

func (s *LearningService) StartLearning(studentID, courseID string) error {
	student, exists := s.storage.GetStudent(studentID)
	if !exists {
		return &storage.NotFoundError{ID: studentID, Type: "student"}
	}

	_, exists = s.storage.GetCourse(courseID)
	if !exists {
		return &storage.NotFoundError{ID: courseID, Type: "course"}
	}

	activeEnrollments := s.getActiveEnrollments(studentID)
	if len(activeEnrollments) >= MaxConcurrentCourses {
		return &MaxCoursesError{Current: len(activeEnrollments), Max: MaxConcurrentCourses}
	}

	completedCourses := s.getCompletedCourses(studentID)
	graph := s.courseService.GetCourseGraph()
	if !graph.CanStartCourse(courseID, completedCourses) {
		return &PrerequisiteNotMetError{CourseID: courseID}
	}

	for _, e := range activeEnrollments {
		if e.CourseID == courseID {
			return nil
		}
	}

	enrollment := &models.Enrollment{
		ID:          utils.GenerateUUID(),
		StudentID:   studentID,
		CourseID:    courseID,
		Status:      "active",
		StartedAt:   time.Now(),
		CompletedAt: nil,
	}

	if err := s.storage.CreateEnrollment(enrollment); err != nil {
		return err
	}

	progress := &models.LearningProgress{
		ID:                 utils.GenerateUUID(),
		EnrollmentID:       enrollment.ID,
		StudentID:          studentID,
		CourseID:           courseID,
		CompletedUnitIDs:   []string{},
		ProgressPercentage: 0,
		UpdatedAt:          time.Now(),
	}

	if err := s.storage.CreateLearningProgress(progress); err != nil {
		return err
	}

	s.achievementSvc.CheckFirstCourse(student.ID)
	s.updateDailyProgress(studentID)

	return nil
}

func (s *LearningService) CompleteUnit(studentID, courseID, unitID string) error {
	progress, exists := s.storage.GetLearningProgress(studentID, courseID)
	if !exists {
		return &storage.NotFoundError{ID: studentID, Type: "learning_progress"}
	}

	course, exists := s.storage.GetCourse(courseID)
	if !exists {
		return &storage.NotFoundError{ID: courseID, Type: "course"}
	}

	var unit *models.Unit
	for i := range course.Units {
		if course.Units[i].ID == unitID {
			unit = &course.Units[i]
			break
		}
	}
	if unit == nil {
		return &storage.NotFoundError{ID: unitID, Type: "unit"}
	}

	if !utils.ContainsString(progress.CompletedUnitIDs, unitID) {
		progress.CompletedUnitIDs = append(progress.CompletedUnitIDs, unitID)
		progress.ProgressPercentage = utils.CalculateProgress(
			len(progress.CompletedUnitIDs),
			len(course.Units),
		)
		progress.UpdatedAt = time.Now()
		s.storage.UpdateLearningProgress(progress)

		record := &models.LearningRecord{
			ID:          utils.GenerateUUID(),
			StudentID:   studentID,
			UnitID:      unitID,
			CourseID:    courseID,
			TimeSpent:   0,
			CompletedAt: time.Now(),
			CreatedAt:   time.Now(),
		}
		s.storage.CreateLearningRecord(record)
	}

	s.updateDailyProgress(studentID)
	s.achievementSvc.CheckConsecutiveDays(studentID)
	s.achievementSvc.CheckTotalHours(studentID)

	return nil
}

func (s *LearningService) GetProgress(studentID, courseID string) (*models.LearningProgress, bool) {
	return s.storage.GetLearningProgress(studentID, courseID)
}

func (s *LearningService) GenerateLearningPath(studentID, targetCourseID string, goal string) (*models.CoursePath, error) {
	student, exists := s.storage.GetStudent(studentID)
	if !exists {
		return nil, &storage.NotFoundError{ID: studentID, Type: "student"}
	}

	completedCourses := s.getCompletedCourses(studentID)
	activeEnrollments := s.getActiveEnrollments(studentID)
	activeCourseIDs := make(map[string]bool)
	for _, e := range activeEnrollments {
		activeCourseIDs[e.CourseID] = true
	}

	graph := s.courseService.GetCourseGraph()
	levels, err := graph.GetLevels()
	if err != nil {
		return nil, err
	}

	requiredCourses := s.collectPrerequisites(targetCourseID, graph, make(map[string]bool))
	requiredCourses[targetCourseID] = true

	pathNodes := []models.PathNode{}
	for courseID := range requiredCourses {
		if completedCourses[courseID] {
			continue
		}

		course, exists := s.storage.GetCourse(courseID)
		if !exists {
			continue
		}

		status := "available"
		progress := 0.0

		if activeCourseIDs[courseID] {
			status = "in_progress"
			if prog, ok := s.storage.GetLearningProgress(studentID, courseID); ok {
				progress = prog.ProgressPercentage
			}
		} else if !graph.CanStartCourse(courseID, completedCourses) {
			status = "locked"
		}

		pathNodes = append(pathNodes, models.PathNode{
			CourseID:      courseID,
			CourseName:    course.Name,
			Level:         levels[courseID],
			ParallelGroup: 0,
			Status:        status,
			Progress:      progress,
		})
	}

	sort.Slice(pathNodes, func(i, j int) bool {
		if pathNodes[i].Level != pathNodes[j].Level {
			return pathNodes[i].Level < pathNodes[j].Level
		}
		return pathNodes[i].CourseID < pathNodes[j].CourseID
	})

	currentLevel := -1
	groupCount := 0
	for i := range pathNodes {
		if pathNodes[i].Level != currentLevel {
			currentLevel = pathNodes[i].Level
			groupCount = 0
		}
		pathNodes[i].ParallelGroup = groupCount
		groupCount++
	}

	path := &models.CoursePath{
		ID:           utils.GenerateUUID(),
		StudentID:    student.ID,
		TargetGoal:   goal,
		PathNodes:    pathNodes,
		TotalCourses: len(pathNodes),
		CreatedAt:    time.Now(),
	}

	return path, nil
}

func (s *LearningService) collectPrerequisites(courseID string, graph *dag.Graph, visited map[string]bool) map[string]bool {
	if visited[courseID] {
		return visited
	}
	visited[courseID] = true

	prereqs := graph.GetPrerequisites(courseID)
	for _, prereqID := range prereqs {
		s.collectPrerequisites(prereqID, graph, visited)
	}

	return visited
}

func (s *LearningService) getActiveEnrollments(studentID string) []*models.Enrollment {
	enrollments := s.storage.GetEnrollmentsByStudent(studentID)
	active := []*models.Enrollment{}
	for _, e := range enrollments {
		if e.Status == "active" {
			active = append(active, e)
		}
	}
	return active
}

func (s *LearningService) getCompletedCourses(studentID string) map[string]bool {
	enrollments := s.storage.GetEnrollmentsByStudent(studentID)
	completed := make(map[string]bool)
	for _, e := range enrollments {
		if e.Status == "completed" {
			completed[e.CourseID] = true
		}
	}
	return completed
}

func (s *LearningService) updateDailyProgress(studentID string) {
	records := s.storage.GetLearningRecordsByStudent(studentID)
	today := time.Now().Truncate(24 * time.Hour)
	for _, r := range records {
		if r.CompletedAt.Truncate(24*time.Hour).Equal(today) {
			return
		}
	}
}

type MaxCoursesError struct {
	Current int
	Max     int
}

func (e *MaxCoursesError) Error() string {
	return "maximum concurrent courses exceeded"
}

type PrerequisiteNotMetError struct {
	CourseID string
}

func (e *PrerequisiteNotMetError) Error() string {
	return "prerequisites not met for course " + e.CourseID
}

func ShuffleQuestions(questions []models.Question, usedQuestions []string) []models.Question {
	rand.Seed(time.Now().UnixNano())

	questionsUsed := make(map[string]bool)
	for _, qid := range usedQuestions {
		questionsUsed[qid] = true
	}

	notUsed := []models.Question{}
	used := []models.Question{}
	for _, q := range questions {
		if questionsUsed[q.ID] {
			used = append(used, q)
		} else {
			notUsed = append(notUsed, q)
		}
	}

	rand.Shuffle(len(notUsed), func(i, j int) {
		notUsed[i], notUsed[j] = notUsed[j], notUsed[i]
	})
	rand.Shuffle(len(used), func(i, j int) {
		used[i], used[j] = used[j], used[i]
	})

	totalNeeded := 10
	minNew := int(float64(totalNeeded) * 0.3)
	if len(notUsed) < minNew {
		minNew = len(notUsed)
	}

	result := []models.Question{}
	result = append(result, notUsed[:minNew]...)
	remaining := totalNeeded - len(result)
	if remaining > 0 && remaining <= len(used) {
		result = append(result, used[:remaining]...)
	}

	rand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})

	return result
}
