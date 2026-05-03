package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"replog"
)

func main() {
	testDir := "./test_log"
	os.RemoveAll(testDir)

	fmt.Println("=== 测试 1: 基本 Append 和 Read ===")
	lg, err := replog.Open(testDir)
	if err != nil {
		log.Fatalf("Open failed: %v", err)
	}

	offset1, err := lg.Append([]byte("message 1"))
	if err != nil {
		log.Fatalf("Append failed: %v", err)
	}
	fmt.Printf("Appended message 1 at offset: %d\n", offset1)

	offset2, err := lg.Append([]byte("message 2"))
	if err != nil {
		log.Fatalf("Append failed: %v", err)
	}
	fmt.Printf("Appended message 2 at offset: %d\n", offset2)

	offset3, err := lg.Append([]byte("message 3"))
	if err != nil {
		log.Fatalf("Append failed: %v", err)
	}
	fmt.Printf("Appended message 3 at offset: %d\n", offset3)

	err = lg.CreateGroup("group1")
	if err != nil {
		log.Fatalf("CreateGroup failed: %v", err)
	}
	fmt.Println("Created consumer group: group1")

	messages, err := lg.Read("group1", 2)
	if err != nil {
		log.Fatalf("Read failed: %v", err)
	}
	fmt.Printf("Read 2 messages from group1: %v\n", messages)

	groupOffset, _ := lg.GroupOffset("group1")
	fmt.Printf("group1 offset after read: %d\n", groupOffset)

	messages, err = lg.Read("group1", 10)
	if err != nil {
		log.Fatalf("Read failed: %v", err)
	}
	fmt.Printf("Read remaining messages from group1: %v\n", messages)

	lg.Close()

	fmt.Println("\n=== 测试 2: 恢复 ===")
	lg2, err := replog.Open(testDir)
	if err != nil {
		log.Fatalf("Open failed: %v", err)
	}

	currentOffset := lg2.CurrentOffset()
	fmt.Printf("Current offset after recovery: %d\n", currentOffset)

	groupOffset, _ = lg2.GroupOffset("group1")
	fmt.Printf("group1 offset after recovery: %d\n", groupOffset)

	lg2.Close()

	fmt.Println("\n=== 测试 3: 截断 ===")
	lg3, err := replog.Open(testDir)
	if err != nil {
		log.Fatalf("Open failed: %v", err)
	}

	err = lg3.Truncate(1)
	if err != nil {
		log.Printf("Truncate(1) error (expected if group offset >= 1): %v", err)
	}

	lg3.CreateGroup("group2")
	lg3.Read("group2", 3)

	fmt.Println("Files before truncate:")
	entries, _ := os.ReadDir(testDir)
	for _, e := range entries {
		info, _ := e.Info()
		fmt.Printf("  %s (%d bytes)\n", e.Name(), info.Size())
	}

	lg3.Append([]byte("message 4"))
	lg3.Append([]byte("message 5"))

	err = lg3.Truncate(3)
	if err != nil {
		fmt.Printf("Truncate(3) error: %v\n", err)
	} else {
		fmt.Println("Truncate(3) succeeded")
	}

	fmt.Println("Files after truncate:")
	entries, _ = os.ReadDir(testDir)
	for _, e := range entries {
		info, _ := e.Info()
		fmt.Printf("  %s (%d bytes)\n", e.Name(), info.Size())
	}

	lg3.Close()

	os.RemoveAll(testDir)

	fmt.Println("\n=== 测试 4: 文件轮转模拟 ===")
	largeTestDir := "./large_test"
	os.RemoveAll(largeTestDir)

	lg4, _ := replog.Open(largeTestDir)
	
	smallEntry := []byte("small")
	for i := 0; i < 100; i++ {
		lg4.Append(append(smallEntry, byte('0'+i%10)))
	}

	fmt.Printf("Current offset: %d\n", lg4.CurrentOffset())
	fmt.Printf("Files created: %d\n", countLogFiles(largeTestDir))

	lg4.Close()
	os.RemoveAll(largeTestDir)

	fmt.Println("\n=== 所有测试完成 ===")
}

func countLogFiles(dir string) int {
	entries, _ := os.ReadDir(dir)
	count := 0
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".offset" {
			count++
		}
	}
	return count
}
