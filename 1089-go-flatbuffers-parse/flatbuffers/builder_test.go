package flatbuffers

import (
	"testing"
)

func TestBuilderSimpleInt32(t *testing.T) {
	builder := NewBuilder(256)
	builder.StartObject(1)
	builder.AddInt32(0, 42, 0)
	obj := builder.EndObject()
	buf := builder.Finish(obj)

	tablePos, err := GetRootAsTable(buf)
	if err != nil {
		t.Fatalf("GetRootAsTable failed: %v", err)
	}

	value, err := GetField(buf, tablePos, 0, TypeInt32)
	if err != nil {
		t.Fatalf("GetField failed: %v", err)
	}

	v, ok := value.(int32)
	if !ok {
		t.Fatalf("Expected int32, got %T", value)
	}

	if v != 42 {
		t.Fatalf("Expected 42, got %d", v)
	}

	t.Logf("Successfully built and read int32 value: %d", v)
}

func TestBuilderMultipleFields(t *testing.T) {
	builder := NewBuilder(256)
	builder.StartObject(4)
	builder.AddInt32(0, 100, 0)
	builder.AddInt64(1, 9999999999, 0)
	builder.AddFloat64(2, 3.14159, 0)
	builder.AddBool(3, true, false)
	obj := builder.EndObject()
	buf := builder.Finish(obj)

	tablePos, err := GetRootAsTable(buf)
	if err != nil {
		t.Fatalf("GetRootAsTable failed: %v", err)
	}

	v0, _ := GetField(buf, tablePos, 0, TypeInt32)
	v1, _ := GetField(buf, tablePos, 1, TypeInt64)
	v2, _ := GetField(buf, tablePos, 2, TypeFloat64)
	v3, _ := GetField(buf, tablePos, 3, TypeBool)

	if v0.(int32) != 100 {
		t.Errorf("Field 0 expected 100, got %v", v0)
	}
	if v1.(int64) != 9999999999 {
		t.Errorf("Field 1 expected 9999999999, got %v", v1)
	}
	if v2.(float64) != 3.14159 {
		t.Errorf("Field 2 expected 3.14159, got %v", v2)
	}
	if v3.(bool) != true {
		t.Errorf("Field 3 expected true, got %v", v3)
	}

	t.Logf("All multi-field tests passed")
}

func TestBuilderString(t *testing.T) {
	builder := NewBuilder(256)
	strOffset := builder.CreateString("Hello FlatBuffers!")
	builder.StartObject(1)
	builder.AddOffset(0, strOffset)
	obj := builder.EndObject()
	buf := builder.Finish(obj)

	tablePos, err := GetRootAsTable(buf)
	if err != nil {
		t.Fatalf("GetRootAsTable failed: %v", err)
	}

	value, err := GetField(buf, tablePos, 0, TypeString)
	if err != nil {
		t.Fatalf("GetField failed: %v", err)
	}

	s, ok := value.(string)
	if !ok {
		t.Fatalf("Expected string, got %T", value)
	}

	if s != "Hello FlatBuffers!" {
		t.Fatalf("Expected 'Hello FlatBuffers!', got '%s'", s)
	}

	t.Logf("Successfully built and read string: %s", s)
}

func TestBuilderMissingField(t *testing.T) {
	builder := NewBuilder(256)
	builder.StartObject(2)
	builder.AddInt32(0, 123, 0)
	obj := builder.EndObject()
	buf := builder.Finish(obj)

	tablePos, err := GetRootAsTable(buf)
	if err != nil {
		t.Fatalf("GetRootAsTable failed: %v", err)
	}

	v0, _ := GetField(buf, tablePos, 0, TypeInt32)
	v1, _ := GetField(buf, tablePos, 1, TypeInt32)

	if v0.(int32) != 123 {
		t.Errorf("Field 0 expected 123, got %v", v0)
	}
	if v1.(int32) != 0 {
		t.Errorf("Field 1 (missing) expected default 0, got %v", v1)
	}

	hasField, _ := HasField(buf, tablePos, 0)
	if !hasField {
		t.Error("Expected HasField(0) to be true")
	}
	hasField, _ = HasField(buf, tablePos, 1)
	if hasField {
		t.Error("Expected HasField(1) to be false")
	}

	t.Logf("Missing field test passed")
}

func TestBuilderNestedTable(t *testing.T) {
	builder := NewBuilder(512)
	
	builder.StartObject(1)
	builder.AddInt32(0, 999, 0)
	nestedObj := builder.EndObject()

	builder.StartObject(2)
	builder.AddOffset(0, nestedObj)
	builder.AddInt32(1, 777, 0)
	rootObj := builder.EndObject()

	buf := builder.Finish(rootObj)

	tablePos, err := GetRootAsTable(buf)
	if err != nil {
		t.Fatalf("GetRootAsTable failed: %v", err)
	}

	v1, _ := GetField(buf, tablePos, 1, TypeInt32)
	if v1.(int32) != 777 {
		t.Errorf("Root field 1 expected 777, got %v", v1)
	}

	nestedPosVal, _ := GetField(buf, tablePos, 0, TypeTable)
	nestedPos := nestedPosVal.(int)

	v0, _ := GetField(buf, nestedPos, 0, TypeInt32)
	if v0.(int32) != 999 {
		t.Errorf("Nested field 0 expected 999, got %v", v0)
	}

	t.Logf("Nested table test passed")
}
