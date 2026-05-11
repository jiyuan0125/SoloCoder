package com.hospital.dto;

import com.hospital.enums.TimeSlot;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDate;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class SlotAvailabilityDTO {
    private Long doctorId;
    private String doctorName;
    private LocalDate appointmentDate;
    private TimeSlot timeSlot;
    private Integer totalSlots;
    private Integer bookedSlots;
    private Integer remainingSlots;
    private Boolean isAvailable;
}
