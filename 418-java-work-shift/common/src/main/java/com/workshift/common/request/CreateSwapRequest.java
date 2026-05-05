package com.workshift.common.request;

import java.time.LocalDate;

public class CreateSwapRequest {
    private String requesterEmployeeId;
    private String targetEmployeeId;
    private LocalDate requesterDate;
    private LocalDate targetDate;
    private String requesterSignature;

    public CreateSwapRequest() {
    }

    public String getRequesterEmployeeId() {
        return requesterEmployeeId;
    }

    public void setRequesterEmployeeId(String requesterEmployeeId) {
        this.requesterEmployeeId = requesterEmployeeId;
    }

    public String getTargetEmployeeId() {
        return targetEmployeeId;
    }

    public void setTargetEmployeeId(String targetEmployeeId) {
        this.targetEmployeeId = targetEmployeeId;
    }

    public LocalDate getRequesterDate() {
        return requesterDate;
    }

    public void setRequesterDate(LocalDate requesterDate) {
        this.requesterDate = requesterDate;
    }

    public LocalDate getTargetDate() {
        return targetDate;
    }

    public void setTargetDate(LocalDate targetDate) {
        this.targetDate = targetDate;
    }

    public String getRequesterSignature() {
        return requesterSignature;
    }

    public void setRequesterSignature(String requesterSignature) {
        this.requesterSignature = requesterSignature;
    }
}