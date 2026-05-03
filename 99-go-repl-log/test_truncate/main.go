package main

import (
	"fmt"
	"os"

	"replog"
)

func main() {
	testDir := "./truncate_test"
	os.RemoveAll(testDir)

	fmt.Println("=== 测试 Truncate 边界情况 ===")
	fmt.Println("场景：Append 5 条数据 (offset 0-4)，CreateGroup 并读完，Truncate(5)")

	// 第一阶段：写入数据并截断
	func() {
		lg, err := replog.Open(testDir)
		if err != nil {
			fmt.Printf("Open failed: %v\n", err)
			os.Exit(1)
		}
		defer lg.Close()

		for i := 0; i < 5; i++ {
			offset, err := lg.Append([]byte(fmt.Sprintf("msg-%d", i)))
			if err != nil {
				fmt.Printf("Append failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Append msg-%d at offset %d\n", i, offset)
		}

		fmt.Printf("\nCurrentOffset: %d\n", lg.CurrentOffset())

		err = lg.CreateGroup("group1")
		if err != nil {
			fmt.Printf("CreateGroup failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Created group1")

		messages, err := lg.Read("group1", 10)
		if err != nil {
			fmt.Printf("Read failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Read from group1: %v\n", messages)

		groupOffset, _ := lg.GroupOffset("group1")
		fmt.Printf("group1 offset after read: %d\n", groupOffset)

		fmt.Println("\n--- Calling Truncate(5) ---")
		err = lg.Truncate(5)
		if err != nil {
			fmt.Printf("Truncate(5) failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Truncate(5) succeeded")

		fmt.Printf("CurrentOffset after truncate: %d\n", lg.CurrentOffset())

		fmt.Println("\n--- Creating new group2 (before restart) ---")
		err = lg.CreateGroup("group2")
		if err != nil {
			fmt.Printf("CreateGroup failed: %v\n", err)
			os.Exit(1)
		}

		group2Offset, _ := lg.GroupOffset("group2")
		fmt.Printf("group2 starting offset: %d\n", group2Offset)

		if group2Offset == -1 {
			fmt.Println("❌  BUG: group2 offset is -1!")
			os.Exit(1)
		} else if group2Offset != 5 {
			fmt.Printf("❌  BUG: expected group2 offset 5, got %d\n", group2Offset)
			os.Exit(1)
		} else {
			fmt.Println("✅  OK: group2 offset is correct (5)")
		}

		fmt.Println("\n--- Append new data after truncate ---")
		newOffset, err := lg.Append([]byte("new-msg"))
		if err != nil {
			fmt.Printf("Append failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Appended new-msg at offset: %d\n", newOffset)

		expectedOffset := int64(5)
		if newOffset != expectedOffset {
			fmt.Printf("❌  BUG: expected offset %d, got %d\n", expectedOffset, newOffset)
			os.Exit(1)
		} else {
			fmt.Println("✅  OK: new offset is correct")
		}
	}()

	// 第二阶段：重启测试
	fmt.Println("\n=== 重启后测试 ===")
	func() {
		lg, err := replog.Open(testDir)
		if err != nil {
			fmt.Printf("Open after restart failed: %v\n", err)
			os.Exit(1)
		}
		defer lg.Close()

		fmt.Printf("CurrentOffset after restart: %d\n", lg.CurrentOffset())

		if lg.CurrentOffset() != 5 {
			fmt.Printf("❌  BUG: expected CurrentOffset 5 after restart, got %d\n", lg.CurrentOffset())
			os.Exit(1)
		} else {
			fmt.Println("✅  OK: CurrentOffset restored correctly")
		}

		group1Offset, _ := lg.GroupOffset("group1")
		fmt.Printf("group1 offset after restart: %d\n", group1Offset)
		if group1Offset != 5 {
			fmt.Printf("❌  BUG: expected group1 offset 5, got %d\n", group1Offset)
			os.Exit(1)
		}

		group2Offset, _ := lg.GroupOffset("group2")
		fmt.Printf("group2 offset after restart: %d\n", group2Offset)

		fmt.Println("\n--- Creating new group3 after restart ---")
		err = lg.CreateGroup("group3")
		if err != nil {
			fmt.Printf("CreateGroup failed: %v\n", err)
			os.Exit(1)
		}

		group3Offset, _ := lg.GroupOffset("group3")
		fmt.Printf("group3 starting offset: %d\n", group3Offset)

		// 现在有数据了，第一个有效 offset 应该是 5
		if group3Offset != 5 {
			fmt.Printf("❌  BUG: expected group3 offset 5, got %d\n", group3Offset)
			os.Exit(1)
		} else {
			fmt.Println("✅  OK: group3 offset is correct")
		}

		fmt.Println("\n--- Reading from group3 ---")
		messages, err := lg.Read("group3", 10)
		if err != nil {
			fmt.Printf("Read failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Read from group3: %v\n", messages)
		if len(messages) == 1 && string(messages[0]) == "new-msg" {
			fmt.Println("✅  OK: Read returned the correct message after restart")
		} else {
			fmt.Println("❌  BUG: Read returned unexpected data")
			os.Exit(1)
		}
	}()

	fmt.Println("\n=== 所有测试通过！===")

	files, _ := os.ReadDir(testDir)
	fmt.Println("Files in test directory:")
	for _, f := range files {
		info, _ := f.Info()
		fmt.Printf("  %s (%d bytes)\n", f.Name(), info.Size())
	}

	os.RemoveAll(testDir)
}
