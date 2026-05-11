package com.chainrestaurant.shiftscheduler.dto;

import lombok.Data;
import lombok.AllArgsConstructor;
import lombok.NoArgsConstructor;

import java.time.LocalDate;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class ComplianceViolation {
    private LocalDate startDate;
    private LocalDate endDate;
    private String message;
    private Integer totalHours;
    private Integer maxAllowedHours;
}
