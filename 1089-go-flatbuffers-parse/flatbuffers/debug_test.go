package flatbuffers

import (
	"fmt"
	"testing"
)

func TestDebugBuilder(t *testing.T) {
	builder := NewBuilder(256)
	builder.StartObject(1)
	builder.AddInt32(0, 42, 0)
	obj := builder.EndObject()
	buf := builder.Finish(obj)

	fmt.Printf("Buffer length: %d\n", len(buf))
	fmt.Printf("Buffer hex: %x\n", buf)

	rootOffset := int(buf[0]) | int(buf[1])<<8 | int(buf[2])<<16 | int(buf[3])<<24
	fmt.Printf("Root offset (LE): %d\n", rootOffset)
	fmt.Printf("Root offset actual read: %d\n", rootOffset)

	if len(buf) >= 12 {
		fmt.Printf("Table at rootOffset:\n")
		for i := 0; i < len(buf); i++ {
			fmt.Printf("  [%d] = 0x%02x\n", i, buf[i])
		}
	}

	tablePos, err := GetRootAsTable(buf)
	if err != nil {
		t.Fatalf("GetRootAsTable failed: %v", err)
	}
	fmt.Printf("tablePos: %d\n", tablePos)

	vtablePos, err := GetTableVOffset(buf, tablePos)
	if err != nil {
		t.Fatalf("GetTableVOffset failed: %v", err)
	}
	fmt.Printf("vtablePos: %d\n", vtablePos)

	fieldOffset, err := GetFieldOffset(buf, tablePos, 0)
	if err != nil {
		t.Fatalf("GetFieldOffset failed: %v", err)
	}
	fmt.Printf("field 0 offset: %d\n", fieldOffset)
}

func TestDebugString(t *testing.T) {
	builder := NewBuilder(256)
	strOffset := builder.CreateString("Hello")
	fmt.Printf("String offset: %d\n", strOffset)
	
	builder.StartObject(1)
	builder.AddOffset(0, strOffset)
	obj := builder.EndObject()
	fmt.Printf("Object offset: %d\n", obj)
	
	buf := builder.Finish(obj)
	
	fmt.Printf("Buffer length: %d\n", len(buf))
	fmt.Printf("Buffer hex: %x\n", buf)
	
	tablePos, err := GetRootAsTable(buf)
	if err != nil {
		t.Fatalf("GetRootAsTable failed: %v", err)
	}
	fmt.Printf("tablePos: %d\n", tablePos)
	
	vtablePos, err := GetTableVOffset(buf, tablePos)
	if err != nil {
		t.Fatalf("GetTableVOffset failed: %v", err)
	}
	fmt.Printf("vtablePos: %d\n", vtablePos)
	
	fieldOffset, err := GetFieldOffset(buf, tablePos, 0)
	if err != nil {
		t.Fatalf("GetFieldOffset failed: %v", err)
	}
	fmt.Printf("field 0 offset: %d\n", fieldOffset)
}

func TestDebugMulti(t *testing.T) {
	builder := NewBuilder(256)
	builder.StartObject(4)
	builder.AddInt32(0, 100, 0)
	builder.AddInt64(1, 9999999999, 0)
	builder.AddFloat64(2, 3.14159, 0)
	builder.AddBool(3, true, false)
	obj := builder.EndObject()
	buf := builder.Finish(obj)

	fmt.Printf("Buffer length: %d\n", len(buf))
	fmt.Printf("Buffer hex: %x\n", buf)

	tablePos, err := GetRootAsTable(buf)
	if err != nil {
		t.Fatalf("GetRootAsTable failed: %v", err)
	}
	fmt.Printf("tablePos: %d\n", tablePos)

	vtablePos, err := GetTableVOffset(buf, tablePos)
	if err != nil {
		t.Fatalf("GetTableVOffset failed: %v", err)
	}
	fmt.Printf("vtablePos: %d\n", vtablePos)

	for i := 0; i < 4; i++ {
		fieldOffset, _ := GetFieldOffset(buf, tablePos, i)
		fmt.Printf("field %d offset: %d\n", i, fieldOffset)
		if fieldOffset != 0 {
			fieldPos := tablePos + int(fieldOffset)
			if fieldPos+8 <= len(buf) {
				fmt.Printf("  fieldPos: %d, hex: %x\n", fieldPos, buf[fieldPos:fieldPos+8])
			}
		}
	}
}

func TestDebugBool(t *testing.T) {
	builder := NewBuilder(256)
	builder.StartObject(1)
	builder.AddBool(0, true, false)
	obj := builder.EndObject()
	buf := builder.Finish(obj)

	fmt.Printf("Buffer length: %d\n", len(buf))
	fmt.Printf("Buffer hex: %x\n", buf)

	tablePos, err := GetRootAsTable(buf)
	if err != nil {
		t.Fatalf("GetRootAsTable failed: %v", err)
	}
	fmt.Printf("tablePos: %d\n", tablePos)

	vtablePos, err := GetTableVOffset(buf, tablePos)
	if err != nil {
		t.Fatalf("GetTableVOffset failed: %v", err)
	}
	fmt.Printf("vtablePos: %d\n", vtablePos)

	fieldOffset, err := GetFieldOffset(buf, tablePos, 0)
	if err != nil {
		t.Fatalf("GetFieldOffset failed: %v", err)
	}
	fmt.Printf("field 0 offset: %d\n", fieldOffset)

	if fieldOffset != 0 {
		fieldPos := tablePos + int(fieldOffset)
		fmt.Printf("fieldPos: %d, value: %d\n", fieldPos, buf[fieldPos])
	}
}
