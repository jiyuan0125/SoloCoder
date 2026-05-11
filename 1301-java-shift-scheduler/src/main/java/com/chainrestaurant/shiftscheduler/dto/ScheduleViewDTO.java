package com.chainrestaurant.shiftscheduler.dto;

import lombok.Data;
import lombok.AllArgsConstructor;
import lombok.NoArgsConstructor;

import java.time.LocalDate;
import java.time.LocalTime;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class ScheduleViewDTO {
    private Long id;
    private Long employeeId;
    private String employeeName;
    private String employeeNo;
    private String storeName;
    private LocalDate scheduleDate;
    private String shiftName;
    private LocalTime shiftStartTime;
    private LocalTime shiftEndTime;
    private Integer shiftDurationHours;
    private Boolean isShiftCrossDay;
    private Boolean isDayOff;
}
