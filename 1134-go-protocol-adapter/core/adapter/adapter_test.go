package adapter

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

type TestUser struct {
	ID        int64     `json:"id" xml:"id" csv:"id"`
	Name      string    `json:"name" xml:"name" csv:"name"`
	Email     string    `json:"email" xml:"email" csv:"email"`
	Age       int       `json:"age" xml:"age" csv:"age"`
	CreatedAt time.Time `json:"created_at" xml:"created_at" csv:"created_at"`
}

func TestCSVDecodeWithQuotes(t *testing.T) {
	adapter := NewCSVAdapter()

	csvData := `"id,name,email,age\n2,Bob,bob@test.com,30"`

	var user TestUser
	if err := adapter.Decode([]byte(csvData), &user); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if user.ID != 2 {
		t.Errorf("Expected ID=2, got %d", user.ID)
	}
	if user.Name != "Bob" {
		t.Errorf("Expected Name='Bob', got '%s'", user.Name)
	}
	if user.Email != "bob@test.com" {
		t.Errorf("Expected Email='bob@test.com', got '%s'", user.Email)
	}
	if user.Age != 30 {
		t.Errorf("Expected Age=30, got %d", user.Age)
	}
}

func TestCSVDecodeNormal(t *testing.T) {
	adapter := NewCSVAdapter()

	csvData := "id,name,email,age\n2,Bob,bob@test.com,30"

	var user TestUser
	if err := adapter.Decode([]byte(csvData), &user); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if user.ID != 2 {
		t.Errorf("Expected ID=2, got %d", user.ID)
	}
	if user.Name != "Bob" {
		t.Errorf("Expected Name='Bob', got '%s'", user.Name)
	}
	if user.Email != "bob@test.com" {
		t.Errorf("Expected Email='bob@test.com', got '%s'", user.Email)
	}
	if user.Age != 30 {
		t.Errorf("Expected Age=30, got %d", user.Age)
	}
}

func TestCSVEncodeMap(t *testing.T) {
	adapter := NewCSVAdapter()

	mapData := map[string]interface{}{
		"id":    1,
		"name":  "Alice",
		"email": "alice@test.com",
		"age":   25,
	}

	result, err := adapter.Encode(mapData)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	resultStr := string(result)

	if strings.Contains(resultStr, "map[") {
		t.Errorf("Result should not contain Go map format, got: %s", resultStr)
	}

	if !strings.Contains(resultStr, "Alice") || !strings.Contains(resultStr, "alice@test.com") {
		t.Errorf("Result should contain CSV data, got: %s", resultStr)
	}

	t.Logf("CSV encode result:\n%s", resultStr)
}

func TestCSVEncodeMapSlice(t *testing.T) {
	adapter := NewCSVAdapter()

	mapSlice := []map[string]interface{}{
		{"id": 1, "name": "Alice", "email": "alice@test.com"},
		{"id": 2, "name": "Bob", "email": "bob@test.com"},
	}

	result, err := adapter.Encode(mapSlice)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	resultStr := string(result)

	if strings.Contains(resultStr, "map[") {
		t.Errorf("Result should not contain Go map format, got: %s", resultStr)
	}

	lines := strings.Split(strings.TrimSpace(resultStr), "\n")
	if len(lines) < 2 {
		t.Errorf("Expected at least 2 lines (header + 1 record), got %d", len(lines))
	}

	t.Logf("CSV encode slice result:\n%s", resultStr)
}

func TestXMLEncodeMap(t *testing.T) {
	adapter := NewXMLAdapter()

	mapData := map[string]interface{}{
		"id":    1,
		"name":  "Alice",
		"email": "alice@test.com",
		"age":   25,
	}

	result, err := adapter.Encode(mapData)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	resultStr := string(result)

	if !strings.Contains(resultStr, "<id>") || !strings.Contains(resultStr, "</id>") {
		t.Errorf("Result should contain valid XML, got: %s", resultStr)
	}

	if strings.Contains(resultStr, "map[") {
		t.Errorf("Result should not contain Go map format, got: %s", resultStr)
	}

	t.Logf("XML encode result:\n%s", resultStr)
}

func TestXMLEncodeMapSlice(t *testing.T) {
	adapter := NewXMLAdapter()

	jsonStr := `[
		{"id": 1, "name": "Alice", "email": "alice@test.com"},
		{"id": 2, "name": "Bob", "email": "bob@test.com"}
	]`

	var mapSlice []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &mapSlice); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	result, err := adapter.Encode(mapSlice)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	resultStr := string(result)

	if strings.Contains(resultStr, "map[") {
		t.Errorf("Result should not contain Go map format, got: %s", resultStr)
	}

	t.Logf("XML encode slice result:\n%s", resultStr)
}

func TestXMLEncodeNestedMap(t *testing.T) {
	adapter := NewXMLAdapter()

	mapData := map[string]interface{}{
		"user": map[string]interface{}{
			"id":    1,
			"name":  "Alice",
		},
		"tags": []string{"admin", "user"},
	}

	result, err := adapter.Encode(mapData)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	resultStr := string(result)

	if strings.Contains(resultStr, "map[") {
		t.Errorf("Result should not contain Go map format, got: %s", resultStr)
	}

	t.Logf("XML encode nested result:\n%s", resultStr)
}
