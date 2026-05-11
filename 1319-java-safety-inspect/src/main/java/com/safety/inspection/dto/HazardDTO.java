package com.safety.inspection.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

import java.time.LocalDateTime;

@Data
public class HazardDTO {

    private Long id;

    @NotNull(message = "区域ID不能为空")
    private Long areaId;

    @NotBlank(message = "隐患等级不能为空")
    private String hazardLevel;

    private String hazardType;

    private String location;

    @NotBlank(message = "隐患描述不能为空")
    private String description;

    private String photos;

    private Long responsiblePersonId;

    private String rectificationPlan;
}
