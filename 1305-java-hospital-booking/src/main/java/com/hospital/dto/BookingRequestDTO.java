package com.hospital.dto;

import com.hospital.enums.TimeSlot;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDate;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class BookingRequestDTO {
    private Long doctorId;
    private LocalDate appointmentDate;
    private TimeSlot timeSlot;
    private String patientName;
    private String patientPhone;
    private String patientIdCard;
}
