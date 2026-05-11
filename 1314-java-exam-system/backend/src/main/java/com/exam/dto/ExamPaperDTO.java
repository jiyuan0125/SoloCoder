package com.exam.dto;

import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

import java.util.List;

@Data
public class ExamPaperDTO {
    private Long id;
    
    @NotBlank(message = "试卷标题不能为空")
    private String title;
    
    private String description;
    
    @NotNull(message = "总分不能为空")
    @Min(value = 1, message = "总分必须大于0")
    private Integer totalScore;
    
    @NotNull(message = "及格分不能为空")
    @Min(value = 0, message = "及格分不能为负数")
    private Integer passScore;
    
    @NotNull(message = "考试时长不能为空")
    @Min(value = 1, message = "考试时长必须大于0")
    private Integer durationMinutes;
    
    private boolean canRetake = false;
    
    private Integer maxRetakeCount;
    
    private Integer switchCountLimit = 3;
    
    private Integer switchDurationLimitSeconds = 30;
    
    private boolean enabled = true;
    
    private List<ExamRuleDTO> rules;
}
