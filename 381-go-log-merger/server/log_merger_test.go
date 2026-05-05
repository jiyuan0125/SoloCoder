package main

import (
	"bufio"
	"os"
	"path/filepath"
	"testing"
)

func TestLogMerger_BasicMerge(t *testing.T) {
	testDir := t.TempDir()

	file1 := filepath.Join(testDir, "file1.log")
	file2 := filepath.Join(testDir, "file2.log")
	output := filepath.Join(testDir, "merged.log")

	os.WriteFile(file1, []byte(`2026-05-04 10:00:00 [INFO] message 1
2026-05-04 10:00:02 [INFO] message 3
2026-05-04 10:00:04 [INFO] message 5
`), 0644)

	os.WriteFile(file2, []byte(`2026-05-04 10:00:01 [INFO] message 2
2026-05-04 10:00:03 [INFO] message 4
2026-05-04 10:00:05 [INFO] message 6
`), 0644)

	merger, err := NewLogMerger([]string{file1, file2}, output, "", false)
	if err != nil {
		t.Fatalf("NewLogMerger failed: %v", err)
	}
	defer merger.Close()

	progressCh := make(chan ProgressUpdate, 10)
	stopCh := make(chan struct{})
	defer close(stopCh)

	err = merger.Merge(progressCh, stopCh)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	content, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("Read output failed: %v", err)
	}

	expected := `2026-05-04 10:00:00 [INFO] message 1
2026-05-04 10:00:01 [INFO] message 2
2026-05-04 10:00:02 [INFO] message 3
2026-05-04 10:00:03 [INFO] message 4
2026-05-04 10:00:04 [INFO] message 5
2026-05-04 10:00:05 [INFO] message 6
`

	if string(content) != expected {
		t.Errorf("Output mismatch.\nExpected:\n%s\nGot:\n%s", expected, string(content))
	}
}

func TestLogMerger_ContinuationLines(t *testing.T) {
	testDir := t.TempDir()

	file1 := filepath.Join(testDir, "file1.log")
	output := filepath.Join(testDir, "merged.log")

	os.WriteFile(file1, []byte(`2026-05-04 10:00:00 [ERROR] exception occurred
java.lang.NullPointerException
    at com.example.Service.doWork(Service.java:42)
2026-05-04 10:00:01 [INFO] recovered
`), 0644)

	os.WriteFile(filepath.Join(testDir, "file2.log"), []byte(`2026-05-04 10:00:00 [INFO] other service
`), 0644)

	merger, err := NewLogMerger([]string{file1, filepath.Join(testDir, "file2.log")}, output, "", false)
	if err != nil {
		t.Fatalf("NewLogMerger failed: %v", err)
	}
	defer merger.Close()

	progressCh := make(chan ProgressUpdate, 10)
	stopCh := make(chan struct{})
	defer close(stopCh)

	err = merger.Merge(progressCh, stopCh)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	lines := readLines(output)
	if len(lines) != 5 {
		t.Errorf("Expected 5 lines, got %d", len(lines))
	}

	if !contains(lines[1], "NullPointerException") {
		t.Errorf("Continuation line not preserved in correct position")
	}
}

func TestLogMerger_GetPositionSeekTo(t *testing.T) {
	testDir := t.TempDir()

	file1 := filepath.Join(testDir, "file1.log")
	content := `2026-05-04 10:00:00 [INFO] message 1
2026-05-04 10:00:01 [INFO] message 2
2026-05-04 10:00:02 [INFO] message 3
`
	os.WriteFile(file1, []byte(content), 0644)

	reader, err := NewLogFileReader(file1, 0, "2006-01-02 15:04:05")
	if err != nil {
		t.Fatalf("NewLogFileReader failed: %v", err)
	}
	defer reader.Close()

	if !reader.HasEntry() {
		t.Fatal("Expected entry")
	}
	_ = reader.NextEntry()
	pos1 := reader.GetPosition()

	if !reader.HasEntry() {
		t.Fatal("Expected second entry")
	}
	entry2 := reader.NextEntry()
	pos2 := reader.GetPosition()

	reader2, err := NewLogFileReader(file1, 0, "2006-01-02 15:04:05")
	if err != nil {
		t.Fatalf("NewLogFileReader failed: %v", err)
	}
	defer reader2.Close()

	err = reader2.SeekTo(pos1)
	if err != nil {
		t.Fatalf("SeekTo failed: %v", err)
	}

	if !reader2.HasEntry() {
		t.Fatal("Expected entry after seek to pos1")
	}
	entryAfterSeek := reader2.NextEntry()

	if entryAfterSeek.Lines[0] != entry2.Lines[0] {
		t.Errorf("SeekTo returned wrong entry.\nExpected: %s\nGot: %s", entry2.Lines[0], entryAfterSeek.Lines[0])
	}

	t.Logf("pos1=%d, pos2=%d", pos1, pos2)
	if pos2 <= pos1 {
		t.Errorf("pos2 should be greater than pos1")
	}
}

func TestLogMerger_ResumeMode(t *testing.T) {
	testDir := t.TempDir()

	file1 := filepath.Join(testDir, "file1.log")
	file2 := filepath.Join(testDir, "file2.log")
	output := filepath.Join(testDir, "merged.log")

	os.WriteFile(file1, []byte(`2026-05-04 10:00:00 [INFO] m1
2026-05-04 10:00:02 [INFO] m3
`), 0644)

	os.WriteFile(file2, []byte(`2026-05-04 10:00:01 [INFO] m2
2026-05-04 10:00:03 [INFO] m4
`), 0644)

	merger1, err := NewLogMerger([]string{file1, file2}, output, "", false)
	if err != nil {
		t.Fatalf("NewLogMerger failed: %v", err)
	}

	progressCh := make(chan ProgressUpdate, 10)
	stopCh := make(chan struct{})

	merger1.readers[0].HasEntry()
	merger1.readers[1].HasEntry()

	ckpt := merger1.CreateCheckpoint()
	t.Logf("Initial checkpoint: OutputPos=%d, Files[0].Position=%d, Files[1].Position=%d",
		ckpt.OutputPos, ckpt.Files[0].Position, ckpt.Files[1].Position)

	merger1.Close()
	close(stopCh)
	close(progressCh)

	firstPassOutput := `2026-05-04 10:00:00 [INFO] m1
2026-05-04 10:00:01 [INFO] m2
`
	os.WriteFile(output, []byte(firstPassOutput), 0644)
	ckpt.OutputPos = int64(len(firstPassOutput))
	ckpt.Files[0].Position = int64(len(`2026-05-04 10:00:00 [INFO] m1
`))
	ckpt.Files[1].Position = int64(len(`2026-05-04 10:00:01 [INFO] m2
`))

	merger2, err := NewLogMerger([]string{file1, file2}, output, "", true)
	if err != nil {
		t.Fatalf("NewLogMerger (resume) failed: %v", err)
	}
	defer merger2.Close()

	merger2.RestoreCheckpoint(ckpt)

	progressCh2 := make(chan ProgressUpdate, 10)
	stopCh2 := make(chan struct{})
	defer close(stopCh2)

	err = merger2.Merge(progressCh2, stopCh2)
	if err != nil {
		t.Fatalf("Merge (resume) failed: %v", err)
	}

	finalContent, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("Read output failed: %v", err)
	}

	expected := `2026-05-04 10:00:00 [INFO] m1
2026-05-04 10:00:01 [INFO] m2
2026-05-04 10:00:02 [INFO] m3
2026-05-04 10:00:03 [INFO] m4
`

	if string(finalContent) != expected {
		t.Errorf("Resume output mismatch.\nExpected:\n%s\nGot:\n%s", expected, string(finalContent))
	}
}

func readLines(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		return nil
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
