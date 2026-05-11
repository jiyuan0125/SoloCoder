package com.safety.inspection.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

import java.time.LocalDate;
import java.util.List;

@Data
public class InspectionPlanDTO {

    private Long id;

    @NotBlank(message = "计划名称不能为空")
    private String planName;

    private String planType;

    @NotNull(message = "区域ID不能为空")
    private Long areaId;

    @NotBlank(message = "巡检频率不能为空")
    private String frequency;

    private Integer frequencyDays;

    @NotNull(message = "开始日期不能为空")
    private LocalDate startDate;

    private LocalDate endDate;

    @NotNull(message = "巡检员ID不能为空")
    private Long inspectorId;

    private String description;

    private Integer status;

    private List<Long> checkItemIds;
}
