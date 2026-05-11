package main

import (
	"fmt"
	"golang.org/x/net/idna"
)

func main() {
	testCases := []string{
		"例子",
		"例",
		"日本語",
		"日本",
	}

	fmt.Println("=== Go 标准库 idna.Lookup ===")
	for _, tc := range testCases {
		encoded, err := idna.Lookup.ToASCII(tc)
		if err != nil {
			fmt.Printf("%q: Error: %v\n", tc, err)
			continue
		}
		fmt.Printf("%q -> %q\n", tc, encoded)
	}

	fmt.Println("\n=== Go 标准库 idna.Display ===")
	for _, tc := range testCases {
		encoded, err := idna.Display.ToASCII(tc)
		if err != nil {
			fmt.Printf("%q: Error: %v\n", tc, err)
			continue
		}
		fmt.Printf("%q -> %q\n", tc, encoded)
	}

	fmt.Println("\n=== Go 标准库 idna.Registration ===")
	for _, tc := range testCases {
		encoded, err := idna.Registration.ToASCII(tc)
		if err != nil {
			fmt.Printf("%q: Error: %v\n", tc, err)
			continue
		}
		fmt.Printf("%q -> %q\n", tc, encoded)
	}
}
