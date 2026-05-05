package com.workshift.common.request;

public class ConfirmSwapRequest {
    private String swapId;
    private String targetEmployeeId;
    private boolean confirmed;
    private String targetSignature;
    private String rejectReason;

    public ConfirmSwapRequest() {
    }

    public String getSwapId() {
        return swapId;
    }

    public void setSwapId(String swapId) {
        this.swapId = swapId;
    }

    public String getTargetEmployeeId() {
        return targetEmployeeId;
    }

    public void setTargetEmployeeId(String targetEmployeeId) {
        this.targetEmployeeId = targetEmployeeId;
    }

    public boolean isConfirmed() {
        return confirmed;
    }

    public void setConfirmed(boolean confirmed) {
        this.confirmed = confirmed;
    }

    public String getTargetSignature() {
        return targetSignature;
    }

    public void setTargetSignature(String targetSignature) {
        this.targetSignature = targetSignature;
    }

    public String getRejectReason() {
        return rejectReason;
    }

    public void setRejectReason(String rejectReason) {
        this.rejectReason = rejectReason;
    }
}