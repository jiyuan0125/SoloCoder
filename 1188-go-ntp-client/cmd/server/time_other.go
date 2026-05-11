//go:build !linux

package main

import "time"

func setSystemTime(t time.Time) (bool, error) {
	return false, nil
}
