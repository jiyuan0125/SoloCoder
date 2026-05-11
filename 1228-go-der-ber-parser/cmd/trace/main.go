package main

import (
	"encoding/json"
	"fmt"

	"der-ber-parser/asn1"
)

func main() {
	jsonStr := `{"type":"INTEGER","value":42}`
	
	var jt asn1.JSONTLV
	err := json.Unmarshal([]byte(jsonStr), &jt)
	if err != nil {
		fmt.Printf("JSON unmarshal error: %v\n", err)
		return
	}

	fmt.Printf("After JSON unmarshal:\n")
	fmt.Printf("  jt.Type = %q (len=%d)\n", jt.Type, len(jt.Type))
	fmt.Printf("  jt.Tag = %q\n", jt.Tag)
	fmt.Printf("  jt.TagNumber = %d\n", jt.TagNumber)
	fmt.Printf("  jt.Constructed = %v\n", jt.Constructed)
	fmt.Printf("  jt.Value = %v (type=%T)\n", jt.Value, jt.Value)
	fmt.Printf("  jt.Children = %v\n", jt.Children)

	// Now trace FromJSON manually
	fmt.Printf("\n=== Tracing FromJSON ===\n")

	fmt.Printf("jt.Type != \"\": %v\n", jt.Type != "")
	
	if jt.Type != "" {
		fmt.Printf("Calling parseTagType(%q)...\n", jt.Type)
	}
}
