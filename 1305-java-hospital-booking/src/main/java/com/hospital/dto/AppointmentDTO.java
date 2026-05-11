package com.hospital.dto;

import com.hospital.enums.AppointmentStatus;
import com.hospital.enums.TimeSlot;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDate;
import java.time.LocalDateTime;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class AppointmentDTO {
    private Long id;
    private Long doctorId;
    private String doctorName;
    private String doctorTitle;
    private Long departmentId;
    private String departmentName;
    private LocalDate appointmentDate;
    private TimeSlot timeSlot;
    private Integer serialNumber;
    private String patientName;
    private String patientPhone;
    private String patientIdCard;
    private AppointmentStatus status;
    private LocalDateTime bookingTime;
    private LocalDateTime cancelTime;
}
