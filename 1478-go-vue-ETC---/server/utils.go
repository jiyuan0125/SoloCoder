package main

import "time"

func parseDate(dateStr string) (time.Time, error) {
	layout := "2006-01-02"
	return time.Parse(layout, dateStr)
}
