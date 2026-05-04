package main

import (
	"business-hours/businesshours"
	"fmt"
	"time"
)

var pass, fail int

func check(name string, got, want interface{}) {
	if got == want {
		pass++
		fmt.Printf("  PASS %s: got %v\n", name, got)
	} else {
		fail++
		fmt.Printf("  FAIL %s: got %v, want %v\n", name, got, want)
	}
}

func main() {
	fmt.Println("=== R2 Verification Tests ===")
	fmt.Println()

	// ---- R1 Bug 1: Cross-midnight Check on next day ----
	fmt.Println("--- R1 Bug 1: Cross-midnight Check ---")
	{
		bh := businesshours.NewBusinessHours()
		bh.SetDailySchedule(time.Monday, &businesshours.DailySchedule{
			OpenSlots: []businesshours.TimeSlot{
				{Start: businesshours.CreateTime(22, 0), End: businesshours.CreateTime(2, 0)},
			},
		})
		// Mon 2024-05-06 is a Monday, Tue 2024-05-07
		r := bh.Check(time.Date(2024, 5, 7, 0, 30, 0, 0, time.Local))
		check("Tue 00:30 in Mon 22-02", r.IsOpen, true)

		r2 := bh.Check(time.Date(2024, 5, 6, 23, 0, 0, 0, time.Local))
		check("Mon 23:00 in Mon 22-02", r2.IsOpen, true)

		r3 := bh.Check(time.Date(2024, 5, 7, 2, 0, 0, 0, time.Local))
		check("Tue 02:00 NOT in Mon 22-02", r3.IsOpen, false)
	}

	// ---- R1 Bug 2: Cross-midnight CalculateHours ----
	fmt.Println("\n--- R1 Bug 2: Cross-midnight CalculateHours ---")
	{
		bh := businesshours.NewBusinessHours()
		bh.SetDailySchedule(time.Monday, &businesshours.DailySchedule{
			OpenSlots: []businesshours.TimeSlot{
				{Start: businesshours.CreateTime(22, 0), End: businesshours.CreateTime(2, 0)},
			},
		})
		r := bh.CalculateHours(
			time.Date(2024, 5, 6, 22, 0, 0, 0, time.Local),
			time.Date(2024, 5, 7, 2, 0, 0, 0, time.Local),
		)
		check("22:00-02:00 = 4h", r.TotalHours, 4.0)
	}

	// ---- R1 Bug 3: Boundary 9:00:00 ----
	fmt.Println("\n--- R1 Bug 3: Boundary check ---")
	{
		bh := businesshours.NewBusinessHours()
		bh.SetDailySchedule(time.Monday, &businesshours.DailySchedule{
			OpenSlots: []businesshours.TimeSlot{
				{Start: businesshours.CreateTime(9, 0), End: businesshours.CreateTime(18, 0)},
			},
		})
		r := bh.Check(time.Date(2024, 5, 6, 9, 0, 0, 0, time.Local))
		check("9:00:00 is open", r.IsOpen, true)

		r2 := bh.Check(time.Date(2024, 5, 6, 18, 0, 0, 0, time.Local))
		check("18:00:00 is NOT open", r2.IsOpen, false)
	}

	// ---- R1 Bug 4: parseWeekday invalid input ----
	// NOTE: parseWeekday is in server/handler.go, not exported.
	// We test via the Config handler indirectly. But we can verify by checking
	// that invalid weekday names are handled. Since parseWeekday is unexported
	// and in the server package, we test the behavior via HTTP.
	// For now, we verify by code review: default case returns (0, nil).
	// This bug is NOT fixable via library import test alone.
	fmt.Println("\n--- R1 Bug 4: parseWeekday invalid input ---")
	fmt.Println("  SKIP (parseWeekday is unexported in server package)")
	fmt.Println("  Code review: default case still returns (0, nil) — NOT FIXED")

	// ---- R1 Bug 5: Client config command ----
	// Client config is now fully implemented (code review confirmed).
	// Cannot test without running server on correct port.
	fmt.Println("\n--- R1 Bug 5: Client config command ---")
	fmt.Println("  SKIP (requires running server)")
	fmt.Println("  Code review: handleConfig() fully implemented — FIXED")

	// ---- Additional: Minute precision ----
	fmt.Println("\n--- Minute precision (8:30-17:30, open 9-17) ---")
	{
		bh := businesshours.NewBusinessHours()
		bh.SetDailySchedule(time.Monday, &businesshours.DailySchedule{
			OpenSlots: []businesshours.TimeSlot{
				{Start: businesshours.CreateTime(9, 0), End: businesshours.CreateTime(17, 0)},
			},
		})
		r := bh.CalculateHours(
			time.Date(2024, 5, 6, 8, 30, 0, 0, time.Local),
			time.Date(2024, 5, 6, 17, 30, 0, 0, time.Local),
		)
		check("8:30-17:30 with 9-17 = 8h", r.TotalHours, 8.0)
	}

	// ---- Additional: Currently open → TimeToOpen = 0 ----
	fmt.Println("\n--- Currently open → TimeToOpen = 0 ---")
	{
		bh := businesshours.NewBusinessHours()
		bh.SetDailySchedule(time.Monday, &businesshours.DailySchedule{
			OpenSlots: []businesshours.TimeSlot{
				{Start: businesshours.CreateTime(9, 0), End: businesshours.CreateTime(18, 0)},
			},
		})
		r := bh.Check(time.Date(2024, 5, 6, 10, 0, 0, 0, time.Local))
		check("IsOpen at 10:00", r.IsOpen, true)
		check("TimeToOpen = 0", r.TimeToOpen, time.Duration(0))
	}

	// ---- Additional: Multiple periods per day ----
	fmt.Println("\n--- Multiple periods per day ---")
	{
		bh := businesshours.NewBusinessHours()
		bh.SetDailySchedule(time.Monday, &businesshours.DailySchedule{
			OpenSlots: []businesshours.TimeSlot{
				{Start: businesshours.CreateTime(7, 0), End: businesshours.CreateTime(9, 0)},
				{Start: businesshours.CreateTime(11, 0), End: businesshours.CreateTime(14, 0)},
			},
		})
		r1 := bh.Check(time.Date(2024, 5, 6, 8, 0, 0, 0, time.Local))
		check("8:00 in 7-9", r1.IsOpen, true)

		r2 := bh.Check(time.Date(2024, 5, 6, 10, 0, 0, 0, time.Local))
		check("10:00 NOT in 7-9 or 11-14", r2.IsOpen, false)

		r3 := bh.CalculateHours(
			time.Date(2024, 5, 6, 7, 0, 0, 0, time.Local),
			time.Date(2024, 5, 6, 14, 0, 0, 0, time.Local),
		)
		check("7-14 with 7-9+11-14 = 5h", r3.TotalHours, 5.0)
	}

	// ---- Additional: Midday break excluded ----
	fmt.Println("\n--- Midday break excluded ---")
	{
		bh := businesshours.NewBusinessHours()
		bh.SetDailySchedule(time.Monday, &businesshours.DailySchedule{
			OpenSlots: []businesshours.TimeSlot{
				{Start: businesshours.CreateTime(9, 0), End: businesshours.CreateTime(17, 0)},
			},
			BreakSlots: []businesshours.TimeSlot{
				{Start: businesshours.CreateTime(12, 0), End: businesshours.CreateTime(13, 0)},
			},
		})
		r := bh.CalculateHours(
			time.Date(2024, 5, 6, 9, 0, 0, 0, time.Local),
			time.Date(2024, 5, 6, 17, 0, 0, 0, time.Local),
		)
		check("9-17 with 12-13 break = 7h", r.TotalHours, 7.0)
	}

	// ---- Additional: Holiday override ----
	fmt.Println("\n--- Holiday override ---")
	{
		bh := businesshours.NewBusinessHours()
		bh.SetDailySchedule(time.Monday, &businesshours.DailySchedule{
			OpenSlots: []businesshours.TimeSlot{
				{Start: businesshours.CreateTime(9, 0), End: businesshours.CreateTime(18, 0)},
			},
		})
		bh.AddHoliday(time.Date(2024, 5, 6, 0, 0, 0, 0, time.Local))
		r := bh.Check(time.Date(2024, 5, 6, 10, 0, 0, 0, time.Local))
		check("Holiday Monday 10:00 NOT open", r.IsOpen, false)
	}

	// ---- Additional: Cross-midnight with break ----
	fmt.Println("\n--- Cross-midnight CalculateHours multi-day ---")
	{
		bh := businesshours.NewBusinessHours()
		bh.SetDailySchedule(time.Friday, &businesshours.DailySchedule{
			OpenSlots: []businesshours.TimeSlot{
				{Start: businesshours.CreateTime(22, 0), End: businesshours.CreateTime(2, 0)},
			},
		})
		// Fri 22:00 to Sat 02:00 = 4h
		r := bh.CalculateHours(
			time.Date(2024, 5, 10, 22, 0, 0, 0, time.Local), // Friday
			time.Date(2024, 5, 11, 2, 0, 0, 0, time.Local),  // Saturday
		)
		check("Fri 22-Sat 02 = 4h", r.TotalHours, 4.0)
	}

	// ---- Summary ----
	fmt.Printf("\n=== Results: %d PASS, %d FAIL ===\n", pass, fail)
	if fail > 0 {
		fmt.Println("OVERALL: FAIL")
	} else {
		fmt.Println("OVERALL: PASS (library tests)")
	}
}
