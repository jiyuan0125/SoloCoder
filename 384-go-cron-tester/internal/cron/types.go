package cron

import (
	"fmt"
	"time"
)

const (
	maxIterations = 10000
)

type FieldType int

const (
	FieldSecond FieldType = iota
	FieldMinute
	FieldHour
	FieldDayOfMonth
	FieldMonth
	FieldDayOfWeek
)

type Field struct {
	Type  FieldType
	Range []int
}

type CronExpr struct {
	Second     *Field
	Minute     *Field
	Hour       *Field
	DayOfMonth *Field
	Month      *Field
	DayOfWeek  *Field
}

type ParseError struct {
	Message string
	Pos     int
	Field   string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s (field: %s, position: %d)", e.Message, e.Field, e.Pos)
}

var fieldNames = map[FieldType]string{
	FieldSecond:     "second",
	FieldMinute:     "minute",
	FieldHour:       "hour",
	FieldDayOfMonth: "day of month",
	FieldMonth:      "month",
	FieldDayOfWeek:  "day of week",
}

var fieldRanges = map[FieldType][2]int{
	FieldSecond:     {0, 59},
	FieldMinute:     {0, 59},
	FieldHour:       {0, 23},
	FieldDayOfMonth: {1, 31},
	FieldMonth:      {1, 12},
	FieldDayOfWeek:  {0, 7},
}

func getWeekday(dow int) time.Weekday {
	if dow == 0 || dow == 7 {
		return time.Sunday
	}
	return time.Weekday(dow)
}

func isLeapYear(year int) bool {
	if year%400 == 0 {
		return true
	}
	if year%100 == 0 {
		return false
	}
	return year%4 == 0
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()
}
