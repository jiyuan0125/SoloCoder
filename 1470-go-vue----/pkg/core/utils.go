package core

import (
	"fmt"
	"math/rand"
	"time"
)

func generateID() string {
	return fmt.Sprintf("%d%06d", time.Now().UnixNano(), rand.Intn(1000000))
}

func now() time.Time {
	return time.Now()
}

func roundToCents(amount float64) float64 {
	return float64(int64(amount*100+0.5)) / 100
}

func formatVisaNumber(counter int) string {
	return fmt.Sprintf("%03d", counter)
}
