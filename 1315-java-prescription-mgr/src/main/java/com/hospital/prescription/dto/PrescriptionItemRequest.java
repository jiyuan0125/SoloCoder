package com.hospital.prescription.dto;

import com.hospital.prescription.enums.DosageForm;
import com.hospital.prescription.enums.SolventType;
import lombok.Data;

import java.math.BigDecimal;

@Data
public class PrescriptionItemRequest {
    private Long drugId;
    private String drugName;
    private DosageForm dosageForm;
    private BigDecimal singleDose;
    private String doseUnit;
    private Integer frequencyPerDay;
    private Integer durationDays;
    private SolventType solventType;
}
