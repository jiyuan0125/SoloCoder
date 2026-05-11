package core

import (
	"testing"
)

const testDSL = `message LoginRequest {
    uint8 version
    uint16 msg_type
    fixed_string(32) username
    fixed_string(32) password
    var_string extra_data
    checksum cs
}`

func TestParseDSL(t *testing.T) {
	def, err := ParseDSL(testDSL)
	if err != nil {
		t.Fatalf("ParseDSL failed: %v", err)
	}

	if def.Name != "LoginRequest" {
		t.Errorf("expected message name 'LoginRequest', got '%s'", def.Name)
	}

	if len(def.Fields) != 6 {
		t.Fatalf("expected 6 fields, got %d", len(def.Fields))
	}

	fieldNames := []string{"version", "msg_type", "username", "password", "extra_data", "cs"}
	for i, name := range fieldNames {
		if def.Fields[i].Name != name {
			t.Errorf("field %d: expected name '%s', got '%s'", i, name, def.Fields[i].Name)
		}
	}

	if def.Fields[0].Type != TypeUint8 {
		t.Errorf("field 0: expected TypeUint8")
	}
	if def.Fields[1].Type != TypeUint16 {
		t.Errorf("field 1: expected TypeUint16")
	}
	if def.Fields[2].Type != TypeFixedString || def.Fields[2].Length != 32 {
		t.Errorf("field 2: expected fixed_string(32)")
	}
	if def.Fields[3].Type != TypeFixedString || def.Fields[3].Length != 32 {
		t.Errorf("field 3: expected fixed_string(32)")
	}
	if def.Fields[4].Type != TypeVarString {
		t.Errorf("field 4: expected TypeVarString")
	}
	if def.Fields[5].Type != TypeChecksum || !def.Fields[5].IsChecksum {
		t.Errorf("field 5: expected checksum")
	}
}

func TestValidateDSL(t *testing.T) {
	if err := ValidateDSL(testDSL); err != nil {
		t.Errorf("ValidateDSL failed for valid DSL: %v", err)
	}

	invalidDSL := `message Test {
		checksum cs
		uint8 version
	}`
	if err := ValidateDSL(invalidDSL); err == nil {
		t.Error("ValidateDSL should fail for checksum not last")
	}

	invalidDSL2 := `message Test {
		uint8 version
		uint8 version
	}`
	if err := ValidateDSL(invalidDSL2); err == nil {
		t.Error("ValidateDSL should fail for duplicate field name")
	}

	invalidDSL3 := `message Test {
		fixed_string(0) name
	}`
	if err := ValidateDSL(invalidDSL3); err == nil {
		t.Error("ValidateDSL should fail for fixed_string(0)")
	}
}

func TestSerializeAndParse(t *testing.T) {
	def, err := ParseDSL(testDSL)
	if err != nil {
		t.Fatalf("ParseDSL failed: %v", err)
	}

	input := map[string]interface{}{
		"version":    1,
		"msg_type":   100,
		"username":   "user123",
		"password":   "pass456",
		"extra_data": "extra info here",
	}

	binaryData, err := SerializeBinary(def, input)
	if err != nil {
		t.Fatalf("SerializeBinary failed: %v", err)
	}

	result, err := ParseBinary(def, binaryData)
	if err != nil {
		t.Fatalf("ParseBinary failed: %v", err)
	}

	if result.Data["version"] != uint8(1) {
		t.Errorf("version mismatch")
	}
	if result.Data["msg_type"] != uint16(100) {
		t.Errorf("msg_type mismatch")
	}
	if result.Data["username"] != "user123" {
		expected := make([]byte, 32)
		copy(expected, "user123")
		if result.Data["username"] != string(expected) {
			t.Errorf("username mismatch")
		}
	}
	if result.Data["password"] != "pass456" {
		expected := make([]byte, 32)
		copy(expected, "pass456")
		if result.Data["password"] != string(expected) {
			t.Errorf("password mismatch")
		}
	}
	if result.Data["extra_data"] != "extra info here" {
		t.Errorf("extra_data mismatch")
	}
}

func TestChecksumVerification(t *testing.T) {
	def, _ := ParseDSL(testDSL)

	input := map[string]interface{}{
		"version":    1,
		"msg_type":   100,
		"username":   "test",
		"password":   "test",
		"extra_data": "test",
	}

	binaryData, _ := SerializeBinary(def, input)
	
	binaryData[len(binaryData)-1]++
	_, err := ParseBinary(def, binaryData)
	if err == nil {
		t.Error("should detect checksum mismatch")
	}
}

func TestVarStringTruncation(t *testing.T) {
	def, _ := ParseDSL(`message Test {
		var_string data
	}`)
	
	binaryData := []byte{0x05, 0x00, 'h', 'e', 'l'}
	_, err := ParseBinary(def, binaryData)
	if err == nil {
		t.Error("should detect truncated var_string data")
	}
}
