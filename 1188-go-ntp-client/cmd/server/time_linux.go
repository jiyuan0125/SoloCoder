package main

import (
	"syscall"
	"time"
)

func setSystemTime(t time.Time) (bool, error) {
	tv := syscall.Timeval{
		Sec:  t.Unix(),
		Usec: int64(t.Nanosecond() / 1000),
	}
	if err := syscall.Settimeofday(&tv); err != nil {
		return false, err
	}
	return true, nil
}
