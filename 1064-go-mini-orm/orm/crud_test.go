package orm

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type TestUser struct {
	ID        int64
	Name      string
	Email     *string
	Age       int
	Active    bool
	CreatedAt time.Time
}

func setupTestDB(t *testing.T) *DB {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite3: %v", err)
	}

	_, err = sqlDB.Exec(`
		CREATE TABLE test_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT,
			age INTEGER,
			active INTEGER,
			created_at TEXT
		)
	`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	return NewDB(sqlDB)
}

func TestInsertAutoIncrement(t *testing.T) {
	db := setupTestDB(t)

	user1 := &TestUser{
		Name:      "Alice",
		Age:       30,
		Active:    true,
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	if err := db.Insert(user1); err != nil {
		t.Fatalf("Insert user1 failed: %v", err)
	}

	user2 := &TestUser{
		Name:      "Bob",
		Age:       25,
		Active:    false,
		CreatedAt: time.Date(2024, 2, 20, 14, 45, 0, 0, time.UTC),
	}

	if err := db.Insert(user2); err != nil {
		t.Fatalf("Insert user2 failed: %v", err)
	}

	t.Log("Both inserts succeeded with auto-increment ID")
}

func TestTimeTimeScan(t *testing.T) {
	db := setupTestDB(t)

	expectedTime := time.Date(2024, 5, 9, 12, 34, 56, 0, time.UTC)

	user := &TestUser{
		Name:      "TestUser",
		Age:       28,
		Active:    true,
		CreatedAt: expectedTime,
	}

	if err := db.Insert(user); err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	var readUser TestUser
	if err := db.GetByID(&readUser, int64(1)); err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if readUser.Name != "TestUser" {
		t.Errorf("Name mismatch: got %q, expected %q", readUser.Name, "TestUser")
	}

	if readUser.Age != 28 {
		t.Errorf("Age mismatch: got %d, expected %d", readUser.Age, 28)
	}

	if !readUser.Active {
		t.Error("Active should be true")
	}

	if !readUser.CreatedAt.Equal(expectedTime) {
		t.Errorf("CreatedAt mismatch: got %v, expected %v", readUser.CreatedAt, expectedTime)
	}

	t.Logf("Successfully read back time.Time: %v", readUser.CreatedAt)
}

func TestList(t *testing.T) {
	db := setupTestDB(t)

	users := []TestUser{
		{Name: "Alice", Age: 30, Active: true, CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "Bob", Age: 25, Active: true, CreatedAt: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "Charlie", Age: 35, Active: false, CreatedAt: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)},
	}

	for i := range users {
		if err := db.Insert(&users[i]); err != nil {
			t.Fatalf("Insert user %d failed: %v", i, err)
		}
	}

	var allUsers []TestUser
	if err := db.List(&allUsers, nil); err != nil {
		t.Fatalf("List all failed: %v", err)
	}

	if len(allUsers) != 3 {
		t.Errorf("Expected 3 users, got %d", len(allUsers))
	}

	var activeUsers []TestUser
	if err := db.List(&activeUsers, &ListOptions{
		Where: map[string]interface{}{"active": 1},
	}); err != nil {
		t.Fatalf("List active failed: %v", err)
	}

	if len(activeUsers) != 2 {
		t.Errorf("Expected 2 active users, got %d", len(activeUsers))
	}

	t.Logf("List test passed: all=%d, active=%d", len(allUsers), len(activeUsers))
}

func TestUpdateAndDelete(t *testing.T) {
	db := setupTestDB(t)

	user := &TestUser{
		Name:      "Original",
		Age:       20,
		Active:    false,
		CreatedAt: time.Now().UTC(),
	}

	if err := db.Insert(user); err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	if err := db.Update(&TestUser{ID: 1}, map[string]interface{}{
		"name":   "Updated",
		"age":    30,
		"active": true,
	}); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	var updated TestUser
	if err := db.GetByID(&updated, int64(1)); err != nil {
		t.Fatalf("GetByID after update failed: %v", err)
	}

	if updated.Name != "Updated" || updated.Age != 30 || !updated.Active {
		t.Errorf("Update not applied correctly: %+v", updated)
	}

	if err := db.Delete(&TestUser{}, int64(1)); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	var shouldNotExist TestUser
	if err := db.GetByID(&shouldNotExist, int64(1)); err == nil {
		t.Error("Expected error for deleted record")
	}

	t.Log("Update and Delete test passed")
}
