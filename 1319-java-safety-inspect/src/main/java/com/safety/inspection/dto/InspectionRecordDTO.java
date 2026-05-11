package com.safety.inspection.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

@Data
public class InspectionRecordDTO {

    private Long id;

    @NotNull(message = "任务ID不能为空")
    private Long taskId;

    @NotNull(message = "检查项ID不能为空")
    private Long checkItemId;

    @NotBlank(message = "检查结果不能为空")
    private String checkResult;

    private String description;

    private String photos;
}
