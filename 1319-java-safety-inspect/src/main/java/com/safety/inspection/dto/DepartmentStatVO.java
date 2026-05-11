package com.safety.inspection.dto;

import lombok.Data;

import java.math.BigDecimal;

@Data
public class DepartmentStatVO {
    private Long departmentId;
    private String departmentName;
    private Integer totalCount;
    private Integer rectifiedCount;
    private BigDecimal rectificationRate;
    private Boolean belowTarget;
}
