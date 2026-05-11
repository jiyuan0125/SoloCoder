package common

import (
	"encoding/json"
	"testing"
)

func TestContainsResponseJSON(t *testing.T) {
	respTrue := ContainsResponse{
		Success:   true,
		Contained: true,
	}

	respFalse := ContainsResponse{
		Success:   true,
		Contained: false,
	}

	dataTrue, err := json.Marshal(respTrue)
	if err != nil {
		t.Fatalf("failed to marshal true response: %v", err)
	}

	dataFalse, err := json.Marshal(respFalse)
	if err != nil {
		t.Fatalf("failed to marshal false response: %v", err)
	}

	t.Logf("contained=true JSON: %s", string(dataTrue))
	t.Logf("contained=false JSON: %s", string(dataFalse))

	var parsedTrue, parsedFalse ContainsResponse
	if err := json.Unmarshal(dataTrue, &parsedTrue); err != nil {
		t.Fatalf("failed to unmarshal true response: %v", err)
	}
	if err := json.Unmarshal(dataFalse, &parsedFalse); err != nil {
		t.Fatalf("failed to unmarshal false response: %v", err)
	}

	if parsedTrue.Contained != true {
		t.Error("expected Contained=true after roundtrip")
	}
	if parsedFalse.Contained != false {
		t.Error("expected Contained=false after roundtrip")
	}

	jsonStrFalse := string(dataFalse)
	expectedContained := `"contained":false`
	if !containsSubstring(jsonStrFalse, expectedContained) {
		t.Errorf("JSON should contain %s, got: %s", expectedContained, jsonStrFalse)
	}
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
