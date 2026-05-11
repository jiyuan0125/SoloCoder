package core

import (
	"errors"
	"fmt"
	"time"

	"github.com/drivingschool/common"
)

const MaxStudentsPerCoach = 8
const MinPracticeInterval = 30 * time.Minute

func (s *Store) AddCoach(name string, teachingType common.VehicleType, phone string) (*common.Coach, error) {
	if name == "" || phone == "" {
		return nil, errors.New("姓名和联系电话不能为空")
	}

	if _, exists := common.VehicleTypeNames[teachingType]; !exists {
		return nil, errors.New("无效的准教车型")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	coachID := fmt.Sprintf("COA_%s_%d", phone, time.Now().Unix())
	coach := &common.Coach{
		ID:              coachID,
		Name:            name,
		TeachingType:    teachingType,
		Phone:           phone,
		CurrentStudents: 0,
		MaxStudents:     MaxStudentsPerCoach,
	}

	s.coaches[coachID] = coach
	s.coachSchedules[coachID] = make(map[string]*common.CoachSchedule)

	return coach, nil
}

func (s *Store) GetCoach(coachID string) (*common.Coach, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	coach, exists := s.coaches[coachID]
	if !exists {
		return nil, errors.New("教练不存在")
	}
	return coach, nil
}

func (s *Store) ListCoaches() []*common.Coach {
	s.mu.RLock()
	defer s.mu.RUnlock()

	coaches := make([]*common.Coach, 0, len(s.coaches))
	for _, coach := range s.coaches {
		coaches = append(coaches, coach)
	}
	return coaches
}

func (s *Store) SetCoachSchedule(coachID string, weekStart time.Time, slots map[time.Weekday][]common.TimeSlot) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.coaches[coachID]; !exists {
		return errors.New("教练不存在")
	}

	scheduleKey := weekStart.Format("2006-01-02")
	schedule := &common.CoachSchedule{
		CoachID:   coachID,
		WeekStart: weekStart,
		Slots:     slots,
	}

	s.coachSchedules[coachID][scheduleKey] = schedule
	return nil
}

func (s *Store) GetCoachSchedule(coachID string, weekStart time.Time) (*common.CoachSchedule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	schedules, exists := s.coachSchedules[coachID]
	if !exists {
		return nil, errors.New("教练不存在")
	}

	scheduleKey := weekStart.Format("2006-01-02")
	schedule, exists := schedules[scheduleKey]
	if !exists {
		return nil, errors.New("该周排班不存在")
	}

	return schedule, nil
}

func (s *Store) BookPractice(studentID, coachID string, timeSlot common.TimeSlot, subject common.Subject) (*common.PracticeBooking, error) {
	if timeSlot.Start.After(timeSlot.End) {
		return nil, errors.New("开始时间必须早于结束时间")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.students[studentID]; !exists {
		return nil, errors.New("学员不存在")
	}

	coach, exists := s.coaches[coachID]
	if !exists {
		return nil, errors.New("教练不存在")
	}

	if coach.CurrentStudents >= coach.MaxStudents {
		return nil, errors.New("教练学员已满")
	}

	weekStart := getWeekStart(timeSlot.Start)
	scheduleKey := weekStart.Format("2006-01-02")
	schedules := s.coachSchedules[coachID]
	schedule, exists := schedules[scheduleKey]
	if !exists {
		return nil, errors.New("教练该周未排班")
	}

	weekday := timeSlot.Start.Weekday()
	daySlots, exists := schedule.Slots[weekday]
	if !exists || len(daySlots) == 0 {
		return nil, errors.New("教练该天不可用")
	}

	isInSlot := false
	for _, slot := range daySlots {
		if !timeSlot.Start.Before(slot.Start) && !timeSlot.End.After(slot.End) {
			isInSlot = true
			break
		}
	}
	if !isInSlot {
		return nil, errors.New("预约时段不在教练可用时间内")
	}

	if err := s.checkPracticeInterval(coachID, timeSlot); err != nil {
		return nil, err
	}

	bookingID := fmt.Sprintf("PBK_%d", time.Now().UnixNano())
	booking := &common.PracticeBooking{
		ID:        bookingID,
		StudentID: studentID,
		CoachID:   coachID,
		TimeSlot:  timeSlot,
		Subject:   subject,
		CreatedAt: time.Now(),
	}

	s.practiceBookings[bookingID] = booking

	student := s.students[studentID]
	if student.CoachID == "" {
		student.CoachID = coachID
		coach.CurrentStudents++
	}

	return booking, nil
}

func (s *Store) checkPracticeInterval(coachID string, newSlot common.TimeSlot) error {
	bookingDate := newSlot.Start.Truncate(24 * time.Hour)

	for _, booking := range s.practiceBookings {
		if booking.CoachID != coachID {
			continue
		}

		bookingDay := booking.TimeSlot.Start.Truncate(24 * time.Hour)
		if !bookingDay.Equal(bookingDate) {
			continue
		}

		if newSlot.Start.Before(booking.TimeSlot.End.Add(MinPracticeInterval)) &&
			newSlot.End.After(booking.TimeSlot.Start.Add(-MinPracticeInterval)) {
			return errors.New("预约时间与现有预约间隔不足30分钟")
		}
	}

	return nil
}

func (s *Store) ListPracticeBookings(coachID string) []*common.PracticeBooking {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bookings := make([]*common.PracticeBooking, 0)
	for _, booking := range s.practiceBookings {
		if coachID == "" || booking.CoachID == coachID {
			bookings = append(bookings, booking)
		}
	}
	return bookings
}

func getWeekStart(t time.Time) time.Time {
	t = t.Truncate(24 * time.Hour)
	offset := int(t.Weekday())
	if offset == 0 {
		offset = 7
	}
	return t.AddDate(0, 0, -(offset - 1))
}
