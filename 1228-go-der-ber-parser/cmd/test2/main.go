package main

import (
	"fmt"

	"der-ber-parser/asn1"
)

// We need to access parseTagType which is unexported.
// Let's test via FromJSON instead.

func main() {
	// Test 1: Direct test of parseTagType behavior via FromJSON
	fmt.Println("=== Test 1: INTEGER ===")
	jt1 := asn1.JSONTLV{
		Type:  "INTEGER",
		Value: float64(42),
	}
	tlv1, err := asn1.FromJSON(jt1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("TLV: Class=%d, Number=%d, Constructed=%v, Value=%v\n",
			tlv1.Class, tlv1.Number, tlv1.Constructed, tlv1.Value)
	}

	// Test 2: Test with lowercase
	fmt.Println("\n=== Test 2: integer (lowercase) ===")
	jt2 := asn1.JSONTLV{
		Type:  "integer",
		Value: float64(42),
	}
	tlv2, err := asn1.FromJSON(jt2)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("TLV: Class=%d, Number=%d, Constructed=%v, Value=%v\n",
			tlv2.Class, tlv2.Number, tlv2.Constructed, tlv2.Value)
	}

	// Test 3: UTF8String
	fmt.Println("\n=== Test 3: UTF8String ===")
	jt3 := asn1.JSONTLV{
		Type:  "UTF8String",
		Value: "Hello",
	}
	tlv3, err := asn1.FromJSON(jt3)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("TLV: Class=%d, Number=%d, Constructed=%v, Value=%v\n",
			tlv3.Class, tlv3.Number, tlv3.Constructed, tlv3.Value)
	}

	// Test 4: Check what happens with empty Type
	fmt.Println("\n=== Test 4: Empty Type ===")
	jt4 := asn1.JSONTLV{}
	tlv4, err := asn1.FromJSON(jt4)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("TLV: Class=%d, Number=%d, Constructed=%v, Value=%v\n",
			tlv4.Class, tlv4.Number, tlv4.Constructed, tlv4.Value)
	}

	// Test 5: Check Encode with the TLVs
	fmt.Println("\n=== Test 5: Encode results ===")
	
	hex1, _ := asn1.EncodeDER(jt1)
	fmt.Printf("INTEGER 42 -> %s (expected: 02012a)\n", hex1)
	
	hex3, _ := asn1.EncodeDER(jt3)
	fmt.Printf("UTF8String Hello -> %s (expected: 0c0548656c6c6f)\n", hex3)
	
	hex4, _ := asn1.EncodeDER(jt4)
	fmt.Printf("Empty -> %s (probably 0000)\n", hex4)
}
