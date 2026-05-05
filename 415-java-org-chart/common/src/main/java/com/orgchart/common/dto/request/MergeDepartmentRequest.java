package com.orgchart.common.dto.request;

public class MergeDepartmentRequest {

    private String targetDepartmentId;

    public MergeDepartmentRequest() {
    }

    public String getTargetDepartmentId() {
        return targetDepartmentId;
    }

    public void setTargetDepartmentId(String targetDepartmentId) {
        this.targetDepartmentId = targetDepartmentId;
    }
}
