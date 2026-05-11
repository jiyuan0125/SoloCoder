package com.safety.inspection.dto;

import lombok.Data;

import java.math.BigDecimal;
import java.util.List;

@Data
public class MonthlyReportVO {
    private Integer reportYear;
    private Integer reportMonth;
    private Integer totalHazards;
    private Integer rectifiedCount;
    private Integer rectifyingCount;
    private Integer overdueCount;
    private Integer generalCount;
    private Integer largerCount;
    private Integer majorCount;
    private BigDecimal rectificationRate;
    private List<DepartmentStatVO> departmentStats;
}
