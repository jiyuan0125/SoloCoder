package com.safety.inspection.entity;

import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import lombok.EqualsAndHashCode;

import java.time.LocalDate;
import java.time.LocalDateTime;

@Data
@EqualsAndHashCode(callSuper = true)
@TableName("inspection_task")
public class InspectionTask extends BaseEntity {

    private String taskNo;

    private Long planId;

    private Long areaId;

    private Long inspectorId;

    private LocalDate taskDate;

    private LocalDateTime startTime;

    private LocalDateTime endTime;

    private String taskStatus;

    private String remark;
}
