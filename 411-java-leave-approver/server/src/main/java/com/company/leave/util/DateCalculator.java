package com.company.leave.util;

import com.company.leave.enums.LeaveType;

import java.time.DayOfWeek;
import java.time.LocalDate;
import java.time.Month;
import java.util.HashSet;
import java.util.Set;

public class DateCalculator {

    private static final Set<LocalDate> HOLIDAYS = new HashSet<>();

    static {
        HOLIDAYS.add(LocalDate.of(2025, Month.JANUARY, 1));
        HOLIDAYS.add(LocalDate.of(2025, Month.JANUARY, 28));
        HOLIDAYS.add(LocalDate.of(2025, Month.JANUARY, 29));
        HOLIDAYS.add(LocalDate.of(2025, Month.JANUARY, 30));
        HOLIDAYS.add(LocalDate.of(2025, Month.JANUARY, 31));
        HOLIDAYS.add(LocalDate.of(2025, Month.FEBRUARY, 1));
        HOLIDAYS.add(LocalDate.of(2025, Month.FEBRUARY, 2));
        HOLIDAYS.add(LocalDate.of(2025, Month.APRIL, 4));
        HOLIDAYS.add(LocalDate.of(2025, Month.MAY, 1));
        HOLIDAYS.add(LocalDate.of(2025, Month.JUNE, 2));
        HOLIDAYS.add(LocalDate.of(2025, Month.OCTOBER, 1));
        HOLIDAYS.add(LocalDate.of(2025, Month.OCTOBER, 2));
        HOLIDAYS.add(LocalDate.of(2025, Month.OCTOBER, 3));
        HOLIDAYS.add(LocalDate.of(2025, Month.OCTOBER, 4));
        HOLIDAYS.add(LocalDate.of(2025, Month.OCTOBER, 5));
        HOLIDAYS.add(LocalDate.of(2025, Month.OCTOBER, 6));
        HOLIDAYS.add(LocalDate.of(2025, Month.OCTOBER, 7));

        HOLIDAYS.add(LocalDate.of(2026, Month.JANUARY, 1));
        HOLIDAYS.add(LocalDate.of(2026, Month.FEBRUARY, 17));
        HOLIDAYS.add(LocalDate.of(2026, Month.FEBRUARY, 18));
        HOLIDAYS.add(LocalDate.of(2026, Month.FEBRUARY, 19));
        HOLIDAYS.add(LocalDate.of(2026, Month.FEBRUARY, 20));
        HOLIDAYS.add(LocalDate.of(2026, Month.FEBRUARY, 21));
        HOLIDAYS.add(LocalDate.of(2026, Month.FEBRUARY, 22));
        HOLIDAYS.add(LocalDate.of(2026, Month.FEBRUARY, 23));
        HOLIDAYS.add(LocalDate.of(2026, Month.APRIL, 5));
        HOLIDAYS.add(LocalDate.of(2026, Month.MAY, 1));
        HOLIDAYS.add(LocalDate.of(2026, Month.JUNE, 20));
        HOLIDAYS.add(LocalDate.of(2026, Month.OCTOBER, 1));
        HOLIDAYS.add(LocalDate.of(2026, Month.OCTOBER, 2));
        HOLIDAYS.add(LocalDate.of(2026, Month.OCTOBER, 3));
        HOLIDAYS.add(LocalDate.of(2026, Month.OCTOBER, 4));
        HOLIDAYS.add(LocalDate.of(2026, Month.OCTOBER, 5));
        HOLIDAYS.add(LocalDate.of(2026, Month.OCTOBER, 6));
        HOLIDAYS.add(LocalDate.of(2026, Month.OCTOBER, 7));
    }

    public static int calculateLeaveDays(LocalDate startDate, LocalDate endDate, LeaveType leaveType) {
        int days = 0;
        LocalDate current = startDate;
        while (!current.isAfter(endDate)) {
            if (isCountedAsLeaveDay(current, leaveType)) {
                days++;
            }
            current = current.plusDays(1);
        }
        return days;
    }

    public static boolean isCountedAsLeaveDay(LocalDate date, LeaveType leaveType) {
        if (isHoliday(date)) {
            return false;
        }
        if (isWeekend(date)) {
            return leaveType.isCountWeekends();
        }
        return true;
    }

    public static boolean isHoliday(LocalDate date) {
        return HOLIDAYS.contains(date);
    }

    public static boolean isWeekend(LocalDate date) {
        DayOfWeek day = date.getDayOfWeek();
        return day == DayOfWeek.SATURDAY || day == DayOfWeek.SUNDAY;
    }
}
