package com.safety.inspection.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

@Data
public class AssignHazardDTO {

    @NotNull(message = "隐患ID不能为空")
    private Long hazardId;

    @NotNull(message = "整改责任人ID不能为空")
    private Long responsiblePersonId;

    @NotBlank(message = "整改方案不能为空")
    private String rectificationPlan;
}
