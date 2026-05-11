//go:build linux && !windows
// +build linux,!windows

//go:generate echo "Generating main..."

package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
