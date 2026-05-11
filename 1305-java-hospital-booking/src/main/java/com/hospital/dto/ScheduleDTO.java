package com.hospital.dto;

import com.hospital.enums.DayOfWeek;
import com.hospital.enums.TimeSlot;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class ScheduleDTO {
    private Long id;
    private Long doctorId;
    private String doctorName;
    private DayOfWeek dayOfWeek;
    private TimeSlot timeSlot;
    private Boolean isAvailable;
}
