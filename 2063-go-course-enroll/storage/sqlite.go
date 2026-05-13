package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type Store interface {
	Close() error
	BeginTx(ctx context.Context) (Tx, error)
	
	CreateCourse(ctx context.Context, tx Tx, course *Course) (int64, error)
	GetCourse(ctx context.Context, id int64) (*Course, error)
	ListCourses(ctx context.Context) ([]*Course, error)
	UpdateCourse(ctx context.Context, tx Tx, course *Course) error
	
	CreateEnrollment(ctx context.Context, tx Tx, enrollment *Enrollment) (int64, error)
	GetEnrollment(ctx context.Context, courseID, studentID int64) (*Enrollment, error)
	ListStudentEnrollments(ctx context.Context, studentID int64) ([]*Enrollment, error)
	ListCourseEnrollments(ctx context.Context, courseID int64) ([]*Enrollment, error)
	UpdateEnrollment(ctx context.Context, tx Tx, enrollment *Enrollment) error
	
	AddToQueue(ctx context.Context, tx Tx, item *WaitQueueItem) (int64, error)
	RemoveFromQueue(ctx context.Context, tx Tx, courseID, studentID int64) error
	GetQueuePosition(ctx context.Context, courseID, studentID int64) (int, error)
	ListQueueItems(ctx context.Context, courseID int64) ([]*WaitQueueItem, error)
	GetQueueLength(ctx context.Context, tx Tx, courseID int64) (int, error)
	PopQueue(ctx context.Context, tx Tx, courseID int64) (*WaitQueueItem, error)
	
	CreateCompletedCourse(ctx context.Context, tx Tx, studentID, courseID int64) error
	HasCompletedCourse(ctx context.Context, studentID, courseID int64) (bool, error)
	GetCompletedCourses(ctx context.Context, studentID int64) ([]int64, error)
	
	CreateStudent(ctx context.Context, tx Tx, student *Student) (int64, error)
	GetStudent(ctx context.Context, id int64) (*Student, error)
	
	CreateResource(ctx context.Context, tx Tx, resource *Resource) error
	GetResource(ctx context.Context, id, resourceType string) (*Resource, error)
	LogResourceAssociation(ctx context.Context, tx Tx, assoc *ResourceAssociation) error
	ListResourceAssociations(ctx context.Context, id, resourceType string) ([]*ResourceAssociation, error)
}

type Tx interface {
	Commit() error
	Rollback() error
}

type Course struct {
	ID            int64
	Name          string
	Capacity      int
	StartTime     time.Time
	EndTime       time.Time
	EnrollStart   time.Time
	EnrollEnd     time.Time
	Prerequisites []int64
	ResourceType  string
	ResourceID    string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Student struct {
	ID   int64
	Name string
}

type Enrollment struct {
	ID         int64
	CourseID   int64
	StudentID  int64
	Status     string
	DropCount  int
	EnrolledAt time.Time
	DroppedAt  time.Time
}

type WaitQueueItem struct {
	ID        int64
	CourseID  int64
	StudentID int64
	Position  int
	CreatedAt time.Time
}

type CompletedCourse struct {
	StudentID int64
	CourseID  int64
	Completed time.Time
}

type Resource struct {
	ID   string
	Type string
	Name string
}

type ResourceAssociation struct {
	ResourceType string
	ResourceID   string
	Operation    string
	EntityType   string
	EntityID     string
	CreatedAt    time.Time
}

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &SQLiteStore{db: db}
	if err := store.init(); err != nil {
		db.Close()
		return nil, err
	}

	return store, nil
}

func (s *SQLiteStore) init() error {
	schema := `
	CREATE TABLE IF NOT EXISTS courses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		capacity INTEGER NOT NULL,
		start_time DATETIME NOT NULL,
		end_time DATETIME NOT NULL,
		enroll_start DATETIME NOT NULL,
		enroll_end DATETIME NOT NULL,
		prerequisites TEXT NOT NULL,
		resource_type TEXT NOT NULL,
		resource_id TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS students (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS enrollments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		course_id INTEGER NOT NULL,
		student_id INTEGER NOT NULL,
		status TEXT NOT NULL,
		drop_count INTEGER NOT NULL DEFAULT 0,
		enrolled_at DATETIME NOT NULL,
		dropped_at DATETIME,
		UNIQUE(course_id, student_id)
	);

	CREATE TABLE IF NOT EXISTS wait_queue (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		course_id INTEGER NOT NULL,
		student_id INTEGER NOT NULL,
		position INTEGER NOT NULL,
		created_at DATETIME NOT NULL,
		UNIQUE(course_id, student_id)
	);

	CREATE TABLE IF NOT EXISTS completed_courses (
		student_id INTEGER NOT NULL,
		course_id INTEGER NOT NULL,
		completed DATETIME NOT NULL,
		PRIMARY KEY (student_id, course_id)
	);

	CREATE TABLE IF NOT EXISTS resources (
		id TEXT NOT NULL,
		type TEXT NOT NULL,
		name TEXT NOT NULL,
		PRIMARY KEY (id, type)
	);

	CREATE TABLE IF NOT EXISTS resource_associations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		resource_type TEXT NOT NULL,
		resource_id TEXT NOT NULL,
		operation TEXT NOT NULL,
		entity_type TEXT NOT NULL,
		entity_id TEXT NOT NULL,
		created_at DATETIME NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_enrollments_course ON enrollments(course_id);
	CREATE INDEX IF NOT EXISTS idx_enrollments_student ON enrollments(student_id);
	CREATE INDEX IF NOT EXISTS idx_queue_course ON wait_queue(course_id);
	CREATE INDEX IF NOT EXISTS idx_assoc_resource ON resource_associations(resource_type, resource_id);
	`

	_, err := s.db.Exec(schema)
	return err
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteStore) BeginTx(ctx context.Context) (Tx, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &sqliteTx{tx: tx}, nil
}

type sqliteTx struct {
	tx *sql.Tx
}

func (t *sqliteTx) Commit() error   { return t.tx.Commit() }
func (t *sqliteTx) Rollback() error { return t.tx.Rollback() }

func encodeInt64Slice(slice []int64) string {
	if slice == nil {
		slice = []int64{}
	}
	b, _ := json.Marshal(slice)
	return string(b)
}

func decodeInt64Slice(s string) []int64 {
	var slice []int64
	if s == "" || s == "null" {
		return []int64{}
	}
	_ = json.Unmarshal([]byte(s), &slice)
	return slice
}

func timeToStr(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func strToTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339Nano, s)
	return t
}
