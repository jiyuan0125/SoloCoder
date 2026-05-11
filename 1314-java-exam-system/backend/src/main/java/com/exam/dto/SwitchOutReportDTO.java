package com.exam.dto;

import jakarta.validation.constraints.NotNull;
import lombok.Data;

@Data
public class SwitchOutReportDTO {
    @NotNull(message = "考试记录ID不能为空")
    private Long examRecordId;
    
    @NotNull(message = "切出时长不能为空")
    private Integer durationSeconds;
}
