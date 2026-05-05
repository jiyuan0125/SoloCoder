package com.employee.common.dto;

import java.util.ArrayList;
import java.util.List;

public class BatchImportResultDTO {
    private int totalCount;
    private int successCount;
    private List<ImportErrorDTO> errors;

    public BatchImportResultDTO() {
        this.errors = new ArrayList<>();
    }

    public int getTotalCount() {
        return totalCount;
    }

    public void setTotalCount(int totalCount) {
        this.totalCount = totalCount;
    }

    public int getSuccessCount() {
        return successCount;
    }

    public void setSuccessCount(int successCount) {
        this.successCount = successCount;
    }

    public List<ImportErrorDTO> getErrors() {
        return errors;
    }

    public void setErrors(List<ImportErrorDTO> errors) {
        this.errors = errors;
    }

    public void addError(int index, String employeeId, String reason) {
        ImportErrorDTO error = new ImportErrorDTO();
        error.setIndex(index);
        error.setEmployeeId(employeeId);
        error.setReason(reason);
        this.errors.add(error);
    }
}
