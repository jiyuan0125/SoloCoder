package com.school.model;

import org.junit.jupiter.api.Test;

import java.time.LocalTime;

import static org.junit.jupiter.api.Assertions.*;

class TimeSlotTest {

    @Test
    void testOverlapsWith_SameTime() {
        TimeSlot slot1 = new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(8, 0), LocalTime.of(9, 40));
        TimeSlot slot2 = new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(8, 0), LocalTime.of(9, 40));
        assertTrue(slot1.overlapsWith(slot2));
    }

    @Test
    void testOverlapsWith_PartialOverlap() {
        TimeSlot slot1 = new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(8, 0), LocalTime.of(9, 40));
        TimeSlot slot2 = new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(9, 30), LocalTime.of(11, 10));
        assertTrue(slot1.overlapsWith(slot2));
    }

    @Test
    void testOverlapsWith_Contained() {
        TimeSlot slot1 = new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(8, 0), LocalTime.of(12, 0));
        TimeSlot slot2 = new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(9, 0), LocalTime.of(10, 0));
        assertTrue(slot1.overlapsWith(slot2));
    }

    @Test
    void testOverlapsWith_NoOverlap() {
        TimeSlot slot1 = new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(8, 0), LocalTime.of(9, 40));
        TimeSlot slot2 = new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(10, 0), LocalTime.of(11, 40));
        assertFalse(slot1.overlapsWith(slot2));
    }

    @Test
    void testOverlapsWith_JustTouching() {
        TimeSlot slot1 = new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(8, 0), LocalTime.of(9, 40));
        TimeSlot slot2 = new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(9, 40), LocalTime.of(11, 20));
        assertTrue(slot1.overlapsWith(slot2));
    }

    @Test
    void testOverlapsWith_DifferentDay() {
        TimeSlot slot1 = new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(8, 0), LocalTime.of(9, 40));
        TimeSlot slot2 = new TimeSlot(TimeSlot.DayOfWeek.TUESDAY, LocalTime.of(8, 0), LocalTime.of(9, 40));
        assertFalse(slot1.overlapsWith(slot2));
    }
}
