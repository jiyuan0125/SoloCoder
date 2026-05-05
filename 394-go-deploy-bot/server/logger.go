package main

import (
	"deploybot/common"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

var (
	logger      *log.Logger
	logFile     *os.File
	currentLevel LogLevel = LevelInfo
)

func initLogger(logPath string, levelStr string) error {
	dir := filepath.Dir(logPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	logFile = f

	switch levelStr {
	case "debug":
		currentLevel = LevelDebug
	case "info":
		currentLevel = LevelInfo
	case "warn":
		currentLevel = LevelWarn
	case "error":
		currentLevel = LevelError
	default:
		currentLevel = LevelInfo
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger = log.New(multiWriter, "", log.LstdFlags)

	return nil
}

func Debug(format string, v ...interface{}) {
	if currentLevel <= LevelDebug {
		logMessage("DEBUG", format, v...)
	}
}

func Info(format string, v ...interface{}) {
	if currentLevel <= LevelInfo {
		logMessage("INFO", format, v...)
	}
}

func Warn(format string, v ...interface{}) {
	if currentLevel <= LevelWarn {
		logMessage("WARN", format, v...)
	}
}

func Error(format string, v ...interface{}) {
	if currentLevel <= LevelError {
		logMessage("ERROR", format, v...)
	}
}

func logMessage(level string, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	logger.Printf("[%s] [%s] %s", timestamp, level, msg)
}

func LogStep(stepName string, action string, duration time.Duration, output string, err error) {
	status := "SUCCESS"
	if err != nil {
		status = "FAILED"
	}

	if duration > 0 {
		logger.Printf("[STEP] [%s] %s - 耗时: %v, 状态: %s", stepName, action, duration, status)
	} else {
		logger.Printf("[STEP] [%s] %s - 状态: %s", stepName, action, status)
	}

	if output != "" {
		logger.Printf("[STEP_OUTPUT] [%s] %s", stepName, output)
	}

	if err != nil {
		logger.Printf("[STEP_ERROR] [%s] %v", stepName, err)
	}
}

func LogSummary(totalDuration time.Duration, steps []*common.StepRecord) {
	logger.Println("========================================")
	logger.Println("          部署总结报告")
	logger.Println("========================================")
	logger.Printf("总耗时: %v\n", totalDuration)
	logger.Println("----------------------------------------")
	logger.Println("步骤耗时分布:")

	for _, step := range steps {
		if step.Duration > 0 {
			status := "✓"
			if step.Status == common.StepFailed {
				status = "✗"
			}
			logger.Printf("  %s [%s] %-20s %v", status, step.Status, step.Name, step.Duration)
		}
	}

	logger.Println("========================================")
}
