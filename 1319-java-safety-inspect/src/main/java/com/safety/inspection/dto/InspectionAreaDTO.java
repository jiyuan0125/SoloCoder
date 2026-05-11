package com.safety.inspection.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
public class InspectionAreaDTO {

    private Long id;

    @NotBlank(message = "区域名称不能为空")
    private String areaName;

    private String areaType;

    private String location;

    private String description;

    private Long managerId;

    private Long departmentId;

    private Integer status;
}
