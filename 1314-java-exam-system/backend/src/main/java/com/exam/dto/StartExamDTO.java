package com.exam.dto;

import jakarta.validation.constraints.NotNull;
import lombok.Data;

@Data
public class StartExamDTO {
    @NotNull(message = "试卷ID不能为空")
    private Long examPaperId;
    
    @NotNull(message = "用户ID不能为空")
    private Long userId;
}
