package main

import (
	"business-hours/businesshours"
	"fmt"
	"time"
)

func main() {
	fmt.Println("=== Test 1: Cross-day business hours ===")
	testCrossDayHours()

	fmt.Println("\n=== Test 2: Boundary check (9:00:00) ===")
	testBoundary()

	fmt.Println("\n=== Test 3: Calculate cross-day hours ===")
	testCalculateCrossDay()

	fmt.Println("\n=== All tests completed ===")
}

func testCrossDayHours() {
	bh := businesshours.NewBusinessHours()

	mondaySchedule := &businesshours.DailySchedule{
		OpenSlots: []businesshours.TimeSlot{
			{Start: businesshours.CreateTime(22, 0), End: businesshours.CreateTime(2, 0)},
		},
	}
	bh.SetDailySchedule(time.Monday, mondaySchedule)

	mondayNight := time.Date(2024, 5, 6, 23, 0, 0, 0, time.Local)
	result := bh.Check(mondayNight)
	fmt.Printf("Monday 23:00 - IsOpen: %v (expected: true)\n", result.IsOpen)

	tuesdayMorning := time.Date(2024, 5, 7, 0, 30, 0, 0, time.Local)
	result2 := bh.Check(tuesdayMorning)
	fmt.Printf("Tuesday 00:30 - IsOpen: %v (expected: true)\n", result2.IsOpen)

	tuesday2am := time.Date(2024, 5, 7, 2, 0, 0, 0, time.Local)
	result3 := bh.Check(tuesday2am)
	fmt.Printf("Tuesday 02:00 - IsOpen: %v (expected: false)\n", result3.IsOpen)
}

func testBoundary() {
	bh := businesshours.NewBusinessHours()

	mondaySchedule := &businesshours.DailySchedule{
		OpenSlots: []businesshours.TimeSlot{
			{Start: businesshours.CreateTime(9, 0), End: businesshours.CreateTime(18, 0)},
		},
	}
	bh.SetDailySchedule(time.Monday, mondaySchedule)

	nineAm := time.Date(2024, 5, 6, 9, 0, 0, 0, time.Local)
	result := bh.Check(nineAm)
	fmt.Printf("Monday 9:00:00 - IsOpen: %v (expected: true)\n", result.IsOpen)

	nineAm1sec := time.Date(2024, 5, 6, 9, 0, 1, 0, time.Local)
	result2 := bh.Check(nineAm1sec)
	fmt.Printf("Monday 9:00:01 - IsOpen: %v (expected: true)\n", result2.IsOpen)

	sixPm := time.Date(2024, 5, 6, 18, 0, 0, 0, time.Local)
	result3 := bh.Check(sixPm)
	fmt.Printf("Monday 18:00:00 - IsOpen: %v (expected: false)\n", result3.IsOpen)
}

func testCalculateCrossDay() {
	bh := businesshours.NewBusinessHours()

	mondaySchedule := &businesshours.DailySchedule{
		OpenSlots: []businesshours.TimeSlot{
			{Start: businesshours.CreateTime(22, 0), End: businesshours.CreateTime(2, 0)},
		},
	}
	bh.SetDailySchedule(time.Monday, mondaySchedule)

	start := time.Date(2024, 5, 6, 22, 0, 0, 0, time.Local)
	end := time.Date(2024, 5, 7, 2, 0, 0, 0, time.Local)
	result := bh.CalculateHours(start, end)
	fmt.Printf("22:00 - 02:00 next day - TotalHours: %.2f (expected: 4.0)\n", result.TotalHours)
	fmt.Printf("Slots count: %d\n", len(result.Slots))
	for i, slot := range result.Slots {
		fmt.Printf("  Slot %d: %s - %s\n", i+1, slot.Start.Format("15:04"), slot.End.Format("15:04"))
	}
}
