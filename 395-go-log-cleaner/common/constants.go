package common

import "time"

const (
	DefaultServerHost = "localhost"
	DefaultServerPort = "8989"
	DefaultFileExtension = ".log"
	DefaultRetainDays = 7
	DefaultCapacityThreshold = 0.8
	DefaultSafeWaterLevel = 0.7
	DefaultMaxDelete = 0
)

func GenerateTaskID() string {
	return time.Now().Format("20060102150405.000000000")
}
