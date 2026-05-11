package com.safety.inspection.dto;

import lombok.Data;

@Data
public class InspectionTaskDTO {
    private Long id;
    private String taskNo;
    private Long planId;
    private Long areaId;
    private Long inspectorId;
    private String taskDate;
    private String taskStatus;
    private String remark;
}
