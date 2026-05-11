package asn1

import (
	"testing"
)

func TestOID2_100(t *testing.T) {
	oid := "2.100"
	
	node, err := NewOID(oid)
	if err != nil {
		t.Fatalf("NewOID(%s) failed: %v", oid, err)
	}

	encoded, err := Encode(node)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	decoded, _, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	resultOID, ok := decoded.AsOID()
	if !ok {
		t.Fatalf("decoded node is not an OID")
	}

	if resultOID != oid {
		t.Errorf("OID mismatch: expected %s, got %s", oid, resultOID)
	}
}

func TestOID1_2_840(t *testing.T) {
	oid := "1.2.840"
	
	node, err := NewOID(oid)
	if err != nil {
		t.Fatalf("NewOID(%s) failed: %v", oid, err)
	}

	encoded, err := Encode(node)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	decoded, _, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	resultOID, ok := decoded.AsOID()
	if !ok {
		t.Fatalf("decoded node is not an OID")
	}

	if resultOID != oid {
		t.Errorf("OID mismatch: expected %s, got %s", oid, resultOID)
	}
}

func TestOID2_5_4_3(t *testing.T) {
	oid := "2.5.4.3"
	
	node, err := NewOID(oid)
	if err != nil {
		t.Fatalf("NewOID(%s) failed: %v", oid, err)
	}

	encoded, err := Encode(node)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	decoded, _, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	resultOID, ok := decoded.AsOID()
	if !ok {
		t.Fatalf("decoded node is not an OID")
	}

	if resultOID != oid {
		t.Errorf("OID mismatch: expected %s, got %s", oid, resultOID)
	}
}
