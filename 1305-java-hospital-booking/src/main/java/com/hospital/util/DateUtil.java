package com.hospital.util;

import com.hospital.enums.DayOfWeek;

import java.time.LocalDate;

public class DateUtil {
    
    public static DayOfWeek toEnumDayOfWeek(java.time.DayOfWeek javaDayOfWeek) {
        return switch (javaDayOfWeek) {
            case MONDAY -> DayOfWeek.MONDAY;
            case TUESDAY -> DayOfWeek.TUESDAY;
            case WEDNESDAY -> DayOfWeek.WEDNESDAY;
            case THURSDAY -> DayOfWeek.THURSDAY;
            case FRIDAY -> DayOfWeek.FRIDAY;
            case SATURDAY -> DayOfWeek.SATURDAY;
            case SUNDAY -> DayOfWeek.SUNDAY;
        };
    }

    public static DayOfWeek getDayOfWeek(LocalDate date) {
        return toEnumDayOfWeek(date.getDayOfWeek());
    }
}
