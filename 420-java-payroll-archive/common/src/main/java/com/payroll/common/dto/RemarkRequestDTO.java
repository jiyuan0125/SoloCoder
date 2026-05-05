package com.payroll.common.dto;

public class RemarkRequestDTO {

    private String remark;

    public RemarkRequestDTO() {
    }

    public RemarkRequestDTO(String remark) {
        this.remark = remark;
    }

    public String getRemark() {
        return remark;
    }

    public void setRemark(String remark) {
        this.remark = remark;
    }
}
