package queue

import (
	"context"
	"fmt"
	"time"

	"course-enroll/storage"
)

type System struct {
	store storage.Store
}

func NewSystem(store storage.Store) *System {
	return &System{store: store}
}

type QueuePosition struct {
	CourseID  int64     `json:"course_id"`
	StudentID int64     `json:"student_id"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
}

func (q *System) AddToQueue(ctx context.Context, tx storage.Tx, courseID, studentID int64) (*QueuePosition, error) {
	item := &storage.WaitQueueItem{
		CourseID:  courseID,
		StudentID: studentID,
	}

	_, err := q.store.AddToQueue(ctx, tx, item)
	if err != nil {
		return nil, fmt.Errorf("add to queue: %w", err)
	}

	return &QueuePosition{
		CourseID:  courseID,
		StudentID: studentID,
		Position:  item.Position,
		CreatedAt: item.CreatedAt,
	}, nil
}

func (q *System) RemoveFromQueue(ctx context.Context, tx storage.Tx, courseID, studentID int64) error {
	return q.store.RemoveFromQueue(ctx, tx, courseID, studentID)
}

func (q *System) GetQueuePosition(ctx context.Context, courseID, studentID int64) (int, error) {
	return q.store.GetQueuePosition(ctx, courseID, studentID)
}

func (q *System) GetQueueLength(ctx context.Context, tx storage.Tx, courseID int64) (int, error) {
	return q.store.GetQueueLength(ctx, tx, courseID)
}

func (q *System) PopFromQueue(ctx context.Context, tx storage.Tx, courseID int64) (*QueuePosition, error) {
	item, err := q.store.PopQueue(ctx, tx, courseID)
	if err != nil {
		return nil, fmt.Errorf("pop queue: %w", err)
	}
	if item == nil {
		return nil, nil
	}

	return &QueuePosition{
		CourseID:  item.CourseID,
		StudentID: item.StudentID,
		Position:  item.Position,
		CreatedAt: item.CreatedAt,
	}, nil
}

func (q *System) ListQueue(ctx context.Context, courseID int64) ([]*QueuePosition, error) {
	items, err := q.store.ListQueueItems(ctx, courseID)
	if err != nil {
		return nil, err
	}

	result := make([]*QueuePosition, len(items))
	for i, item := range items {
		result[i] = &QueuePosition{
			CourseID:  item.CourseID,
			StudentID: item.StudentID,
			Position:  item.Position,
			CreatedAt: item.CreatedAt,
		}
	}
	return result, nil
}
