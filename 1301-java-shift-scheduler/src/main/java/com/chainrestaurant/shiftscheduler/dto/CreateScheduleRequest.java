package com.chainrestaurant.shiftscheduler.dto;

import lombok.Data;
import java.time.LocalDate;

@Data
public class CreateScheduleRequest {
    private Long employeeId;
    private LocalDate scheduleDate;
    private Long shiftId;
    private Boolean isDayOff;
}
