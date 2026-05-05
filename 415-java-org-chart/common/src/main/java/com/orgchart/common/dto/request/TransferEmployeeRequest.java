package com.orgchart.common.dto.request;

public class TransferEmployeeRequest {

    private String newDepartmentId;
    private String reason;

    public TransferEmployeeRequest() {
    }

    public String getNewDepartmentId() {
        return newDepartmentId;
    }

    public void setNewDepartmentId(String newDepartmentId) {
        this.newDepartmentId = newDepartmentId;
    }

    public String getReason() {
        return reason;
    }

    public void setReason(String reason) {
        this.reason = reason;
    }
}
