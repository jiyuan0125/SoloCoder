package com.safety.inspection.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

@Data
public class InspectionCheckItemDTO {

    private Long id;

    @NotNull(message = "区域ID不能为空")
    private Long areaId;

    @NotBlank(message = "检查项名称不能为空")
    private String itemName;

    private String itemDescription;

    private String standard;

    private String riskLevel;

    private Integer sort;

    private Integer status;
}
