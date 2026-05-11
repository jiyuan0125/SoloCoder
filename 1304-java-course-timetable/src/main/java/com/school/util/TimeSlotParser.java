package com.school.util;

import com.school.model.TimeSlot;

import java.time.LocalTime;
import java.time.format.DateTimeFormatter;

public class TimeSlotParser {
    private static final DateTimeFormatter TIME_FORMATTER = DateTimeFormatter.ofPattern("HH:mm");

    public static TimeSlot parse(String timeSlotString) {
        String[] parts = timeSlotString.trim().split("\\s+");
        if (parts.length != 2) {
            throw new IllegalArgumentException("无效的时段格式：" + timeSlotString + "，正确格式如：周一 08:00-09:40");
        }

        String dayStr = parts[0];
        String timeRange = parts[1];

        TimeSlot.DayOfWeek dayOfWeek = TimeSlot.DayOfWeek.fromDisplayName(dayStr);

        String[] timeParts = timeRange.split("-");
        if (timeParts.length != 2) {
            throw new IllegalArgumentException("无效的时间范围格式：" + timeRange + "，正确格式如：08:00-09:40");
        }

        LocalTime startTime = LocalTime.parse(timeParts[0].trim(), TIME_FORMATTER);
        LocalTime endTime = LocalTime.parse(timeParts[1].trim(), TIME_FORMATTER);

        if (startTime.isAfter(endTime)) {
            throw new IllegalArgumentException("开始时间不能晚于结束时间");
        }

        return new TimeSlot(dayOfWeek, startTime, endTime);
    }

    public static String format(TimeSlot timeSlot) {
        return timeSlot.toString();
    }
}
