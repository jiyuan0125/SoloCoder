package com.safety.inspection.entity;

import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import lombok.EqualsAndHashCode;

import java.time.LocalDate;

@Data
@EqualsAndHashCode(callSuper = true)
@TableName("inspection_plan")
public class InspectionPlan extends BaseEntity {

    private String planName;

    private String planType;

    private Long areaId;

    private String frequency;

    private Integer frequencyDays;

    private LocalDate startDate;

    private LocalDate endDate;

    private Long inspectorId;

    private String description;

    private Integer status;

    private Long createdBy;
}
