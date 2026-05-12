package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port   string
	DBPath string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "medical.db"
	}

	return &Config{
		Port:   port,
		DBPath: dbPath,
	}
}

func IsWorkDay(t time.Time) bool {
	weekday := t.Weekday()
	return weekday >= time.Monday && weekday <= time.Friday
}

func IsWorkTime(t time.Time) bool {
	if !IsWorkDay(t) {
		return false
	}
	hour := t.Hour()
	minute := t.Minute()
	if hour < 8 || (hour == 8 && minute < 30) {
		return false
	}
	if hour > 17 || (hour == 17 && minute > 30) {
		return false
	}
	return true
}

func GetNextWorkDayStart(t time.Time) time.Time {
	for {
		t = t.AddDate(0, 0, 1)
		if IsWorkDay(t) {
			return time.Date(t.Year(), t.Month(), t.Day(), 8, 30, 0, 0, t.Location())
		}
	}
}

func CalculateResponseStartTime(reportTime time.Time) time.Time {
	if IsWorkTime(reportTime) {
		return reportTime
	}
	workStart := time.Date(reportTime.Year(), reportTime.Month(), reportTime.Day(), 8, 30, 0, 0, reportTime.Location())
	if reportTime.Before(workStart) && IsWorkDay(reportTime) {
		return workStart
	}
	return GetNextWorkDayStart(reportTime)
}
