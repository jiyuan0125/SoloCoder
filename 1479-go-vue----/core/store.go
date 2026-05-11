package core

import (
	"sync"

	"github.com/drivingschool/common"
)

type Store struct {
	mu sync.RWMutex

	students        map[string]*common.Student
	coaches         map[string]*common.Coach
	coachSchedules  map[string]map[string]*common.CoachSchedule
	practiceBookings map[string]*common.PracticeBooking
	examPlans       map[string]*common.ExamPlan
	examBookings    map[string]*common.ExamBooking
	pendingCancellations map[string]struct{}
}

func NewStore() *Store {
	return &Store{
		students:             make(map[string]*common.Student),
		coaches:              make(map[string]*common.Coach),
		coachSchedules:       make(map[string]map[string]*common.CoachSchedule),
		practiceBookings:     make(map[string]*common.PracticeBooking),
		examPlans:            make(map[string]*common.ExamPlan),
		examBookings:         make(map[string]*common.ExamBooking),
		pendingCancellations: make(map[string]struct{}),
	}
}
