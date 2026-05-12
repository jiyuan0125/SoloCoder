package diagnostic

import (
	"fmt"
	"log"
	"os"
	"time"

	"protocol-converter/config"
)

var (
	logger    *log.Logger
	logFile   *os.File
)

func init() {
	logPath := os.Getenv("DIAGNOSTIC_LOG")
	if logPath == "" {
		logPath = "diagnostic.log"
	}

	var err error
	logFile, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger = log.New(os.Stderr, "DIAGNOSTIC: ", log.LstdFlags)
	} else {
		logger = log.New(logFile, "", 0)
	}
}

func LogUnknownField(path, fieldName string) {
	lock := config.GetWriteLock()
	lock.Lock()
	defer lock.Unlock()

	timestamp := time.Now().Format(time.RFC3339)
	message := fmt.Sprintf("[%s] UNKNOWN_FIELD path=%s field=%s\n", timestamp, path, fieldName)

	if _, err := logger.Writer().Write([]byte(message)); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write diagnostic log: %v\n", err)
	}
}

func LogConversionError(path, fieldName string, err error) {
	lock := config.GetWriteLock()
	lock.Lock()
	defer lock.Unlock()

	timestamp := time.Now().Format(time.RFC3339)
	message := fmt.Sprintf("[%s] CONVERSION_ERROR path=%s field=%s error=%v\n", timestamp, path, fieldName, err)

	if _, writeErr := logger.Writer().Write([]byte(message)); writeErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to write diagnostic log: %v\n", writeErr)
	}
}

func Close() {
	if logFile != nil {
		logFile.Close()
	}
}
