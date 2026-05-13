package enrollment

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"course-enroll/modules/course"
	"course-enroll/modules/queue"
	"course-enroll/modules/schedule"
	"course-enroll/storage"
)

var (
	ErrNotEnrollmentPeriod   = errors.New("not in enrollment period")
	ErrCourseFull            = errors.New("course is full")
	ErrPrerequisitesNotMet   = errors.New("prerequisites not met")
	ErrAlreadyEnrolled       = errors.New("already enrolled")
	ErrAlreadyInQueue        = errors.New("already in wait queue")
	ErrDropPeriodExpired     = errors.New("drop period expired")
	ErrCourseNotStarted      = errors.New("course not started")
	ErrNotEnrolled           = errors.New("not enrolled")
)

type Engine struct {
	store        storage.Store
	courseMgr    *course.Manager
	queueSys     *queue.System
	scheduleChk  *schedule.Checker
	notifyFunc   func(studentID int64, message string)
}

func NewEngine(store storage.Store, courseMgr *course.Manager, queueSys *queue.System, scheduleChk *schedule.Checker) *Engine {
	return &Engine{
		store:       store,
		courseMgr:   courseMgr,
		queueSys:    queueSys,
		scheduleChk: scheduleChk,
	}
}

func (e *Engine) SetNotificationFunc(fn func(studentID int64, message string)) {
	e.notifyFunc = fn
}

type EnrollResult struct {
	Status   string `json:"status"`
	Message  string `json:"message,omitempty"`
	Position int    `json:"position,omitempty"`
}

type MissingPrerequisite struct {
	CourseID   int64  `json:"course_id"`
	CourseName string `json:"course_name"`
}

type EnrollError struct {
	Code        string
	Status      int
	Message     string
	MissingPrereqs []MissingPrerequisite
	Conflicts   []schedule.Conflict
}

func (e *EnrollError) Error() string { return e.Message }

func (e *Engine) Enroll(ctx context.Context, studentID, courseID int64) (*EnrollResult, *EnrollError) {
	now := time.Now()

	courseInfo, err := e.store.GetCourse(ctx, courseID)
	if err != nil {
		return nil, &EnrollError{Status: 500, Message: err.Error()}
	}
	if courseInfo == nil {
		return nil, &EnrollError{Code: "NOT_FOUND", Status: 404, Message: "course not found"}
	}

	if now.Before(courseInfo.EnrollStart) || now.After(courseInfo.EnrollEnd) {
		return nil, &EnrollError{Code: "OUTSIDE_PERIOD", Status: 403, Message: "outside enrollment period"}
	}

	existingEnroll, err := e.store.GetEnrollment(ctx, courseID, studentID)
	if err != nil {
		return nil, &EnrollError{Status: 500, Message: err.Error()}
	}
	if existingEnroll != nil && existingEnroll.Status == "enrolled" {
		return nil, &EnrollError{Code: "DUPLICATE", Status: 409, Message: "already enrolled"}
	}

	queuePos, err := e.queueSys.GetQueuePosition(ctx, courseID, studentID)
	if err != nil {
		return nil, &EnrollError{Status: 500, Message: err.Error()}
	}
	if queuePos > 0 {
		return nil, &EnrollError{Code: "IN_QUEUE", Status: 409, Message: "already in wait queue"}
	}

	missing, err := e.checkPrerequisites(ctx, studentID, courseInfo.Prerequisites)
	if err != nil {
		return nil, &EnrollError{Status: 500, Message: err.Error()}
	}
	if len(missing) > 0 {
		return nil, &EnrollError{
			Code:           "MISSING_PREREQ",
			Status:         400,
			Message:        "prerequisites not met",
			MissingPrereqs: missing,
		}
	}

	conflicts, err := e.scheduleChk.CheckConflict(ctx, studentID, courseInfo)
	if err != nil {
		return nil, &EnrollError{Status: 500, Message: err.Error()}
	}
	if len(conflicts) > 0 {
		return nil, &EnrollError{
			Code:      "SCHEDULE_CONFLICT",
			Status:    409,
			Message:   "schedule conflict",
			Conflicts: conflicts,
		}
	}

	enrolledCount, err := e.courseMgr.GetEnrolledCount(ctx, courseID)
	if err != nil {
		return nil, &EnrollError{Status: 500, Message: err.Error()}
	}

	tx, err := e.store.BeginTx(ctx)
	if err != nil {
		return nil, &EnrollError{Status: 500, Message: err.Error()}
	}
	defer tx.Rollback()

	if enrolledCount < courseInfo.Capacity {
		result, enrollErr := e.doEnroll(ctx, tx, studentID, courseInfo)
		if enrollErr != nil {
			return nil, enrollErr
		}

		if err := tx.Commit(); err != nil {
			return nil, &EnrollError{Status: 500, Message: err.Error()}
		}
		return result, nil
	}

	queueItem, qErr := e.queueSys.AddToQueue(ctx, tx, courseID, studentID)
	if qErr != nil {
		return nil, &EnrollError{Status: 500, Message: qErr.Error()}
	}

	if courseInfo.ResourceType != "" && courseInfo.ResourceID != "" {
		err = e.store.LogResourceAssociation(ctx, tx, &storage.ResourceAssociation{
			ResourceType: courseInfo.ResourceType,
			ResourceID:   courseInfo.ResourceID,
			Operation:    "queue",
			EntityType:   "student",
			EntityID:     strconv.FormatInt(studentID, 10),
		})
		if err != nil {
			return nil, &EnrollError{Status: 500, Message: err.Error()}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, &EnrollError{Status: 500, Message: err.Error()}
	}

	return &EnrollResult{
		Status:   "queued",
		Message:  fmt.Sprintf("course is full, added to wait queue at position %d", queueItem.Position),
		Position: queueItem.Position,
	}, nil
}

func (e *Engine) doEnroll(ctx context.Context, tx storage.Tx, studentID int64, courseInfo *storage.Course) (*EnrollResult, *EnrollError) {
	enrollment := &storage.Enrollment{
		CourseID:  courseInfo.ID,
		StudentID: studentID,
		Status:    "enrolled",
	}

	_, err := e.store.CreateEnrollment(ctx, tx, enrollment)
	if err != nil {
		return nil, &EnrollError{Status: 500, Message: err.Error()}
	}

	if courseInfo.ResourceType != "" && courseInfo.ResourceID != "" {
		err = e.store.LogResourceAssociation(ctx, tx, &storage.ResourceAssociation{
			ResourceType: courseInfo.ResourceType,
			ResourceID:   courseInfo.ResourceID,
			Operation:    "enroll",
			EntityType:   "student",
			EntityID:     strconv.FormatInt(studentID, 10),
		})
		if err != nil {
			return nil, &EnrollError{Status: 500, Message: err.Error()}
		}
	}

	return &EnrollResult{Status: "enrolled", Message: "enrollment successful"}, nil
}

func (e *Engine) checkPrerequisites(ctx context.Context, studentID int64, prereqs []int64) ([]MissingPrerequisite, error) {
	if len(prereqs) == 0 {
		return nil, nil
	}

	completed, err := e.store.GetCompletedCourses(ctx, studentID)
	if err != nil {
		return nil, err
	}

	completedMap := make(map[int64]bool)
	for _, id := range completed {
		completedMap[id] = true
	}

	var missing []MissingPrerequisite
	for _, prereqID := range prereqs {
		if !completedMap[prereqID] {
			prereqCourse, err := e.store.GetCourse(ctx, prereqID)
			if err != nil {
				return nil, err
			}
			name := fmt.Sprintf("Course %d", prereqID)
			if prereqCourse != nil {
				name = prereqCourse.Name
			}
			missing = append(missing, MissingPrerequisite{
				CourseID:   prereqID,
				CourseName: name,
			})
		}
	}

	return missing, nil
}

type DropResult struct {
	Status        string `json:"status"`
	DropCount     int    `json:"drop_count,omitempty"`
	NotifiedQueue bool   `json:"notified_queue,omitempty"`
}

func (e *Engine) Drop(ctx context.Context, studentID, courseID int64) (*DropResult, *EnrollError) {
	now := time.Now()

	courseInfo, err := e.store.GetCourse(ctx, courseID)
	if err != nil {
		return nil, &EnrollError{Status: 500, Message: err.Error()}
	}
	if courseInfo == nil {
		return nil, &EnrollError{Code: "NOT_FOUND", Status: 404, Message: "course not found"}
	}

	enrollment, err := e.store.GetEnrollment(ctx, courseID, studentID)
	if err != nil {
		return nil, &EnrollError{Status: 500, Message: err.Error()}
	}
	if enrollment == nil || enrollment.Status != "enrolled" {
		return nil, &EnrollError{Code: "NOT_ENROLLED", Status: 400, Message: "not enrolled"}
	}

	oneWeekAfterStart := courseInfo.StartTime.Add(7 * 24 * time.Hour)
	
	needsRecord := now.After(courseInfo.StartTime)
	if now.After(oneWeekAfterStart) {
		return nil, &EnrollError{Code: "DROP_EXPIRED", Status: 400, Message: "drop period expired"}
	}

	tx, err := e.store.BeginTx(ctx)
	if err != nil {
		return nil, &EnrollError{Status: 500, Message: err.Error()}
	}
	defer tx.Rollback()

	enrollment.Status = "dropped"
	enrollment.DroppedAt = now
	if needsRecord {
		enrollment.DropCount++
	}

	if err := e.store.UpdateEnrollment(ctx, tx, enrollment); err != nil {
		return nil, &EnrollError{Status: 500, Message: err.Error()}
	}

	if courseInfo.ResourceType != "" && courseInfo.ResourceID != "" {
		err = e.store.LogResourceAssociation(ctx, tx, &storage.ResourceAssociation{
			ResourceType: courseInfo.ResourceType,
			ResourceID:   courseInfo.ResourceID,
			Operation:    "drop",
			EntityType:   "student",
			EntityID:     strconv.FormatInt(studentID, 10),
		})
		if err != nil {
			return nil, &EnrollError{Status: 500, Message: err.Error()}
		}
	}

	notified := false
	nextStudent, err := e.queueSys.PopFromQueue(ctx, tx, courseID)
	if err != nil {
		return nil, &EnrollError{Status: 500, Message: err.Error()}
	}
	if nextStudent != nil {
		nextEnrollment := &storage.Enrollment{
			CourseID:  courseID,
			StudentID: nextStudent.StudentID,
			Status:    "enrolled",
		}

		_, err = e.store.CreateEnrollment(ctx, tx, nextEnrollment)
		if err != nil {
			return nil, &EnrollError{Status: 500, Message: err.Error()}
		}

		if courseInfo.ResourceType != "" && courseInfo.ResourceID != "" {
			err = e.store.LogResourceAssociation(ctx, tx, &storage.ResourceAssociation{
				ResourceType: courseInfo.ResourceType,
				ResourceID:   courseInfo.ResourceID,
				Operation:    "enroll_from_queue",
				EntityType:   "student",
				EntityID:     strconv.FormatInt(nextStudent.StudentID, 10),
			})
			if err != nil {
				return nil, &EnrollError{Status: 500, Message: err.Error()}
			}
		}

		notified = true
		if e.notifyFunc != nil {
			e.notifyFunc(nextStudent.StudentID, fmt.Sprintf("You have been enrolled in course %d from wait queue", courseID))
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, &EnrollError{Status: 500, Message: err.Error()}
	}

	result := &DropResult{
		Status:        "dropped",
		DropCount:     enrollment.DropCount,
		NotifiedQueue: notified,
	}

	return result, nil
}

func (e *Engine) CreateStudent(ctx context.Context, name string) (int64, error) {
	tx, err := e.store.BeginTx(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	id, err := e.store.CreateStudent(ctx, tx, &storage.Student{Name: name})
	if err != nil {
		return 0, err
	}

	return id, tx.Commit()
}

func (e *Engine) MarkCourseCompleted(ctx context.Context, studentID, courseID int64) error {
	tx, err := e.store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := e.store.CreateCompletedCourse(ctx, tx, studentID, courseID); err != nil {
		return err
	}

	return tx.Commit()
}
