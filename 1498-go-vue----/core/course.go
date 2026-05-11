package core

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"housekeeping-training/common"
)

func (s *Store) CreateCourse(name string, level common.CourseLevel, hours int, instructor string, fee int, maxCapacity int) (*common.Course, error) {
	if name == "" {
		return nil, errors.New("course name is required")
	}
	if level != common.LevelBeginner && level != common.LevelAdvanced {
		return nil, errors.New("invalid course level")
	}
	if hours <= 0 {
		return nil, errors.New("hours must be positive")
	}
	if fee < 0 {
		return nil, errors.New("fee cannot be negative")
	}
	if maxCapacity <= 0 {
		return nil, errors.New("max capacity must be positive")
	}

	s.Lock()
	defer s.Unlock()

	s.nextCourseID++
	course := &common.Course{
		ID:          fmt.Sprintf("C%04d", s.nextCourseID),
		Name:        name,
		Level:       level,
		Hours:       hours,
		Instructor:  instructor,
		Fee:         fee,
		MaxCapacity: maxCapacity,
	}
	s.courses[course.ID] = course
	return course, nil
}

func (s *Store) GetCourse(id string) (*common.Course, error) {
	s.RLock()
	defer s.RUnlock()

	course, ok := s.courses[id]
	if !ok {
		return nil, errors.New("course not found")
	}
	return course, nil
}

func (s *Store) ListCourses() []*common.Course {
	s.RLock()
	defer s.RUnlock()

	result := make([]*common.Course, 0, len(s.courses))
	for _, c := range s.courses {
		result = append(result, c)
	}
	return result
}

func (s *Store) CreateSchedule(courseID, startDateStr, timeSlot, classroom string, totalClasses int) (*common.ClassSchedule, error) {
	if courseID == "" {
		return nil, errors.New("course ID is required")
	}
	if startDateStr == "" {
		return nil, errors.New("start date is required")
	}
	if timeSlot == "" {
		return nil, errors.New("time slot is required")
	}
	if classroom == "" {
		return nil, errors.New("classroom is required")
	}
	if totalClasses <= 0 {
		return nil, errors.New("total classes must be positive")
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return nil, errors.New("invalid start date format, use YYYY-MM-DD")
	}

	s.Lock()
	defer s.Unlock()

	if _, ok := s.courses[courseID]; !ok {
		return nil, errors.New("course not found")
	}

	s.nextScheduleID++
	schedule := &common.ClassSchedule{
		ID:           fmt.Sprintf("S%04d", s.nextScheduleID),
		CourseID:     courseID,
		StartDate:    startDate,
		TimeSlot:     timeSlot,
		Classroom:    classroom,
		Enrolled:     0,
		TotalClasses: totalClasses,
	}
	s.schedules[schedule.ID] = schedule
	return schedule, nil
}

func (s *Store) GetSchedule(id string) (*common.ClassSchedule, error) {
	s.RLock()
	defer s.RUnlock()

	schedule, ok := s.schedules[id]
	if !ok {
		return nil, errors.New("schedule not found")
	}
	return schedule, nil
}

func (s *Store) ListSchedules() []*common.ClassSchedule {
	s.RLock()
	defer s.RUnlock()

	result := make([]*common.ClassSchedule, 0, len(s.schedules))
	for _, sc := range s.schedules {
		result = append(result, sc)
	}
	return result
}

func (s *Store) InitDefaultCourses() {
	defaultCourses := []struct {
		name       string
		level      common.CourseLevel
		hours      int
		instructor string
		fee        int
		capacity   int
	}{
		{"家政基础知识", common.LevelBeginner, 20, "张老师", 500, 30},
		{"清洁技能", common.LevelBeginner, 30, "李老师", 800, 25},
		{"烹饪技能", common.LevelBeginner, 40, "王老师", 1000, 20},
		{"安全常识", common.LevelBeginner, 15, "赵老师", 400, 35},
		{"母婴护理", common.LevelAdvanced, 60, "陈老师", 2000, 15},
		{"老人护理", common.LevelAdvanced, 50, "刘老师", 1800, 18},
		{"营养配餐", common.LevelAdvanced, 45, "孙老师", 1500, 20},
		{"急救知识", common.LevelAdvanced, 25, "周老师", 900, 25},
	}

	for _, c := range defaultCourses {
		s.CreateCourse(c.name, c.level, c.hours, c.instructor, c.fee, c.capacity)
	}
}

func (s *Store) CheckScheduleCapacity(scheduleID string) bool {
	s.RLock()
	defer s.RUnlock()

	schedule, ok := s.schedules[scheduleID]
	if !ok {
		return false
	}
	course, ok := s.courses[schedule.CourseID]
	if !ok {
		return false
	}
	return schedule.Enrolled < course.MaxCapacity
}

func (s *Store) IncrementEnrollment(scheduleID string) error {
	s.Lock()
	defer s.Unlock()

	schedule, ok := s.schedules[scheduleID]
	if !ok {
		return errors.New("schedule not found")
	}
	course, ok := s.courses[schedule.CourseID]
	if !ok {
		return errors.New("course not found")
	}
	if schedule.Enrolled >= course.MaxCapacity {
		return errors.New("schedule is full")
	}
	schedule.Enrolled++
	return nil
}

func (s *Store) DecrementEnrollment(scheduleID string) error {
	s.Lock()
	defer s.Unlock()

	schedule, ok := s.schedules[scheduleID]
	if !ok {
		return errors.New("schedule not found")
	}
	if schedule.Enrolled > 0 {
		schedule.Enrolled--
	}
	return nil
}

func ParseDateStr(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

func (s *Store) getCourseID(id string) int {
	if len(id) <= 1 {
		return 0
	}
	num, _ := strconv.Atoi(id[1:])
	return num
}
