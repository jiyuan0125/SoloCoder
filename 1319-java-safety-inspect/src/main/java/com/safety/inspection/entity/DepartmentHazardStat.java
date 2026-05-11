package com.safety.inspection.entity;

import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.math.BigDecimal;
import java.time.LocalDateTime;

@Data
@TableName("department_hazard_stat")
public class DepartmentHazardStat {

    private Long id;

    private Long reportId;

    private Long departmentId;

    private Integer totalCount;

    private Integer rectifiedCount;

    private BigDecimal rectificationRate;

    private LocalDateTime createdAt;
}
