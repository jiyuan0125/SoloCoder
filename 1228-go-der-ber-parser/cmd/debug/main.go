package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"der-ber-parser/asn1"
)

func main() {
	// Test 1: Direct FromJSON + Encode
	fmt.Println("=== Test 1: Direct FromJSON + Encode ===")
	jt1 := asn1.JSONTLV{
		Type:  "INTEGER",
		Value: float64(42),
	}
	fmt.Printf("Input JSONTLV: Type=%s, Value=%v\n", jt1.Type, jt1.Value)

	tlv1, err := asn1.FromJSON(jt1)
	if err != nil {
		fmt.Printf("FromJSON error: %v\n", err)
	} else {
		fmt.Printf("TLV: Class=%d, Number=%d, Constructed=%v, Value=%v\n",
			tlv1.Class, tlv1.Number, tlv1.Constructed, tlv1.Value)
	}

	bytes1, err := asn1.Encode(tlv1, asn1.DefaultEncodeOptions(asn1.ModeDER))
	if err != nil {
		fmt.Printf("Encode error: %v\n", err)
	} else {
		fmt.Printf("Encoded: %s\n", hex.EncodeToString(bytes1))
	}

	// Test 2: EncodeDER
	fmt.Println("\n=== Test 2: EncodeDER ===")
	hex2, err := asn1.EncodeDER(jt1)
	if err != nil {
		fmt.Printf("EncodeDER error: %v\n", err)
	} else {
		fmt.Printf("EncodeDER result: %s\n", hex2)
	}

	// Test 3: Parse JSON from string
	fmt.Println("\n=== Test 3: Parse JSON from string ===")
	jsonStr := `{"type":"INTEGER","value":42}`
	var jt3 asn1.JSONTLV
	err = json.Unmarshal([]byte(jsonStr), &jt3)
	if err != nil {
		fmt.Printf("JSON unmarshal error: %v\n", err)
	} else {
		fmt.Printf("Parsed JSONTLV: Type=%s, Tag=%s, TagNumber=%d, Value=%v\n",
			jt3.Type, jt3.Tag, jt3.TagNumber, jt3.Value)

		hex3, err := asn1.EncodeDER(jt3)
		if err != nil {
			fmt.Printf("EncodeDER error: %v\n", err)
		} else {
			fmt.Printf("EncodeDER result: %s\n", hex3)
		}
	}

	// Test 4: UTF8String
	fmt.Println("\n=== Test 4: UTF8String ===")
	jt4 := asn1.JSONTLV{
		Type:  "UTF8String",
		Value: "Hello",
	}
	hex4, err := asn1.EncodeDER(jt4)
	if err != nil {
		fmt.Printf("EncodeDER error: %v\n", err)
	} else {
		fmt.Printf("UTF8String Hello: %s\n", hex4)
	}

	// Test 5: What happens with empty JSONTLV?
	fmt.Println("\n=== Test 5: Empty JSONTLV ===")
	jt5 := asn1.JSONTLV{}
	fmt.Printf("Empty JSONTLV: Type=%s, Tag=%s, TagNumber=%d, Constructed=%v\n",
		jt5.Type, jt5.Tag, jt5.TagNumber, jt5.Constructed)

	tlv5, err := asn1.FromJSON(jt5)
	if err != nil {
		fmt.Printf("FromJSON error: %v\n", err)
	} else {
		fmt.Printf("TLV: Class=%d, Number=%d, Constructed=%v\n",
			tlv5.Class, tlv5.Number, tlv5.Constructed)
		bytes5, _ := asn1.Encode(tlv5, asn1.DefaultEncodeOptions(asn1.ModeDER))
		fmt.Printf("Encoded: %s\n", hex.EncodeToString(bytes5))
	}
}
