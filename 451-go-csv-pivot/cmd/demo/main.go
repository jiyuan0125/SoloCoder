// Package main demonstrates the CSV join library.
// This program shows how to use the csvjoin package directly.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"csvjoin/pkg/csvjoin"
)

func main() {
	fmt.Println("=== CSV Join Library Demo ===")
	fmt.Println()

	// Create a temporary directory for demo files
	tempDir := "./demo_data"
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	// Create sample CSV files for demonstration
	createSampleFiles(tempDir)

	// Demo 1: Basic left join
	fmt.Println("Demo 1: Basic Left Join")
	fmt.Println(strings.Repeat("-", 40))

	joinOpts := csvjoin.JoinOptions{
		Files: []csvjoin.FileOptions{
			{
				Path:       filepath.Join(tempDir, "users.csv"),
				SkipHeader: false,
				FileSuffix: "_user",
			},
			{
				Path:       filepath.Join(tempDir, "orders.csv"),
				SkipHeader: false,
				FileSuffix: "_order",
			},
		},
		JoinKey:  "user_id",
		JoinType: csvjoin.LeftJoin,
		TrimKeys: true,
	}

	joiner, err := csvjoin.NewJoiner(joinOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create joiner: %v\n", err)
		os.Exit(1)
	}

	outputPath := filepath.Join(tempDir, "result_left.csv")
	result, err := joiner.Join(outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Join failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Output: %s\n", result.OutputPath)
	fmt.Printf("Row count: %d\n", result.RowCount)
	fmt.Printf("Columns: %v\n", result.Columns)
	fmt.Println()

	// Display the result
	displayCSV(outputPath)

	// Demo 2: Inner join
	fmt.Println("\nDemo 2: Inner Join")
	fmt.Println(strings.Repeat("-", 40))

	joinOpts.JoinType = csvjoin.InnerJoin
	joiner2, _ := csvjoin.NewJoiner(joinOpts)
	outputPath2 := filepath.Join(tempDir, "result_inner.csv")
	result2, _ := joiner2.Join(outputPath2)

	fmt.Printf("Output: %s\n", result2.OutputPath)
	fmt.Printf("Row count: %d\n", result2.RowCount)
	fmt.Printf("Columns: %v\n", result2.Columns)
	fmt.Println()

	displayCSV(outputPath2)

	// Demo 3: Three files join
	fmt.Println("\nDemo 3: Join Three Files")
	fmt.Println(strings.Repeat("-", 40))

	joinOpts3 := csvjoin.JoinOptions{
		Files: []csvjoin.FileOptions{
			{
				Path:       filepath.Join(tempDir, "users.csv"),
				SkipHeader: false,
				FileSuffix: "_u",
			},
			{
				Path:       filepath.Join(tempDir, "orders.csv"),
				SkipHeader: false,
				FileSuffix: "_o",
			},
			{
				Path:       filepath.Join(tempDir, "products.csv"),
				SkipHeader: false,
				FileSuffix: "_p",
			},
		},
		JoinKey:  "user_id",
		JoinType: csvjoin.LeftJoin,
		TrimKeys: true,
	}

	joiner3, _ := csvjoin.NewJoiner(joinOpts3)
	outputPath3 := filepath.Join(tempDir, "result_three.csv")
	result3, _ := joiner3.Join(outputPath3)

	fmt.Printf("Output: %s\n", result3.OutputPath)
	fmt.Printf("Row count: %d\n", result3.RowCount)
	fmt.Printf("Columns: %v\n", result3.Columns)
	fmt.Println()

	displayCSV(outputPath3)

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println()
	fmt.Println("Key Features Demonstrated:")
	fmt.Println("  ✓ RFC 4180 compliant CSV parsing")
	fmt.Println("  ✓ Left join (all left rows preserved)")
	fmt.Println("  ✓ Inner join (only matching rows)")
	fmt.Println("  ✓ Multiple file joins (3+ files)")
	fmt.Println("  ✓ Automatic column name suffixing")
	fmt.Println("  ✓ Key value trimming")
	fmt.Println("  ✓ Memory-efficient processing")
}

// createSampleFiles creates sample CSV files for the demo.
func createSampleFiles(tempDir string) {
	// Users CSV
	usersContent := `user_id,name,email
1,Alice,alice@example.com
2,Bob,bob@example.com
3,Charlie,charlie@example.com
4,Diana,diana@example.com
`
	os.WriteFile(filepath.Join(tempDir, "users.csv"), []byte(usersContent), 0644)

	// Orders CSV (note: user_id 5 has no matching user, user 4 has no orders)
	ordersContent := `order_id,user_id,product_id,amount
101,1,P001,99.99
102,1,P002,49.50
103,2,P001,99.99
104,3,P003,150.00
105,5,P001,99.99
`
	os.WriteFile(filepath.Join(tempDir, "orders.csv"), []byte(ordersContent), 0644)

	// Products CSV
	productsContent := `product_id,user_id,product_name,category
P001,1,Laptop,Electronics
P002,1,Mouse,Accessories
P003,2,Keyboard,Accessories
P004,3,Monitor,Electronics
`
	os.WriteFile(filepath.Join(tempDir, "products.csv"), []byte(productsContent), 0644)

	fmt.Println("Sample CSV files created:")
	fmt.Println("  - users.csv (4 users)")
	fmt.Println("  - orders.csv (5 orders, user 5 has no matching user)")
	fmt.Println("  - products.csv (4 products)")
	fmt.Println()
}

// displayCSV displays the contents of a CSV file.
func displayCSV(filePath string) {
	rows, columns, err := csvjoin.ReadAllRows(filePath, false, nil)
	if err != nil {
		fmt.Printf("Error reading CSV: %v\n", err)
		return
	}

	fmt.Println("Columns:", columns)
	fmt.Println("Rows:")
	for i, row := range rows {
		fmt.Printf("  %d: %v\n", i+1, row)
	}
}
