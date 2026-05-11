package com.hospital.prescription.dto;

import lombok.Builder;
import lombok.Data;

import java.util.List;

@Data
@Builder
public class PrescriptionValidationResult {
    private boolean canSubmit;
    private boolean hasBlockingErrors;
    private boolean hasWarnings;
    private List<ValidationResult> validations;
    private boolean requiresOverDoseConfirmation;
    private String confirmationMessage;
    
    public List<ValidationResult> getBlockingValidations() {
        return validations.stream()
                .filter(ValidationResult::isBlocking)
                .toList();
    }
    
    public List<ValidationResult> getWarningValidations() {
        return validations.stream()
                .filter(v -> !v.isBlocking())
                .toList();
    }
}
