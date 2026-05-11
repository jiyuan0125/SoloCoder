package com.hospital.prescription.dto;

import lombok.Data;

import java.util.List;

@Data
public class PrescriptionRequest {
    private Long patientId;
    private String diagnosis;
    private String doctorName;
    private String department;
    private List<PrescriptionItemRequest> items;
    private boolean overDoseConfirmed;
}
