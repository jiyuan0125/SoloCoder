package com.school.model;

import java.time.LocalTime;
import java.util.Objects;

public class TimeSlot {
    private DayOfWeek dayOfWeek;
    private LocalTime startTime;
    private LocalTime endTime;

    public TimeSlot() {
    }

    public TimeSlot(DayOfWeek dayOfWeek, LocalTime startTime, LocalTime endTime) {
        this.dayOfWeek = dayOfWeek;
        this.startTime = startTime;
        this.endTime = endTime;
    }

    public DayOfWeek getDayOfWeek() {
        return dayOfWeek;
    }

    public void setDayOfWeek(DayOfWeek dayOfWeek) {
        this.dayOfWeek = dayOfWeek;
    }

    public LocalTime getStartTime() {
        return startTime;
    }

    public void setStartTime(LocalTime startTime) {
        this.startTime = startTime;
    }

    public LocalTime getEndTime() {
        return endTime;
    }

    public void setEndTime(LocalTime endTime) {
        this.endTime = endTime;
    }

    public boolean overlapsWith(TimeSlot other) {
        if (this.dayOfWeek != other.dayOfWeek) {
            return false;
        }
        return !this.startTime.isAfter(other.endTime) && !other.startTime.isAfter(this.endTime);
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        TimeSlot timeSlot = (TimeSlot) o;
        return dayOfWeek == timeSlot.dayOfWeek &&
                Objects.equals(startTime, timeSlot.startTime) &&
                Objects.equals(endTime, timeSlot.endTime);
    }

    @Override
    public int hashCode() {
        return Objects.hash(dayOfWeek, startTime, endTime);
    }

    @Override
    public String toString() {
        return dayOfWeek.getDisplayName() + " " + startTime + "-" + endTime;
    }

    public enum DayOfWeek {
        MONDAY("周一"),
        TUESDAY("周二"),
        WEDNESDAY("周三"),
        THURSDAY("周四"),
        FRIDAY("周五"),
        SATURDAY("周六"),
        SUNDAY("周日");

        private final String displayName;

        DayOfWeek(String displayName) {
            this.displayName = displayName;
        }

        public String getDisplayName() {
            return displayName;
        }

        public static DayOfWeek fromDisplayName(String displayName) {
            for (DayOfWeek day : values()) {
                if (day.displayName.equals(displayName)) {
                    return day;
                }
            }
            throw new IllegalArgumentException("Unknown day: " + displayName);
        }

        public int getOrder() {
            return ordinal();
        }
    }
}
