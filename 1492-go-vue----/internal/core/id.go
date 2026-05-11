package core

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"
)

var counter int64

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GenerateID(prefix string) string {
	timestamp := time.Now().Unix()
	cnt := atomic.AddInt64(&counter, 1)
	random := rand.Intn(10000)
	return fmt.Sprintf("%s-%d-%06d-%04d", prefix, timestamp, cnt, random)
}
