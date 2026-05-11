// Package testpkg is an example package for testing the Go doc generator.
//
// This package demonstrates various Go features that should be properly
// documented by the generator, including:
//   - Structs with fields and tags
//   - Interfaces with method signatures
//   - Functions with parameters and return values
//   - Methods on structs
//   - Constants and variables
//
// Example:
//
//	package main
//	
//	import "testpkg"
//	
//	func main() {
//	    cfg := testpkg.NewConfig("test")
//	    fmt.Println(cfg.Name)
//	}
package testpkg
