package com.hospital.prescription.dto;

import com.hospital.prescription.enums.SeverityLevel;
import lombok.Builder;
import lombok.Data;

@Data
@Builder
public class ValidationResult {
    private String validationType;
    private String message;
    private SeverityLevel severity;
    private String relatedDrugNames;
    private String suggestion;
    
    public boolean isBlocking() {
        return severity != null && severity.isMustBlock();
    }
}
