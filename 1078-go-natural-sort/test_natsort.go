//go:build ignore

package main

import (
	"fmt"
	"naturalsort/pkg/natsort"
)

func main() {
	testCases := [][]string{
		{"file10", "file1", "file2"},
		{"file001", "file1", "file01"},
		{"-3", "1", "-5", "0"},
		{"1.5", "1.15", "1.2"},
		{"Apple", "apple", "Banana"},
		{"中文2", "中文10", "中文1"},
		{"file１", "file10", "file２"},
	}

	for i, tc := range testCases {
		opts := natsort.DefaultOptions()
		sorted := natsort.Sort(tc, opts)
		fmt.Printf("Test %d: %v => %v\n", i+1, tc, sorted)
	}

	opts := natsort.DefaultOptions()
	opts.IgnoreLeadingZeros = false
	leadingZerosTest := []string{"file001", "file1", "file01"}
	sorted := natsort.Sort(leadingZerosTest, opts)
	fmt.Printf("\nTest (keep leading zeros): %v => %v\n", leadingZerosTest, sorted)

	opts2 := natsort.DefaultOptions()
	opts2.Ascending = false
	descTest := []string{"file1", "file10", "file2"}
	descSorted := natsort.Sort(descTest, opts2)
	fmt.Printf("\nTest (descending): %v => %v\n", descTest, descSorted)
}
