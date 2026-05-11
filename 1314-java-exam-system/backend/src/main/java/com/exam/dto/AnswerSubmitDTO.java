package com.exam.dto;

import jakarta.validation.constraints.NotNull;
import lombok.Data;

@Data
public class AnswerSubmitDTO {
    @NotNull(message = "考试记录ID不能为空")
    private Long examRecordId;
    
    @NotNull(message = "考试题目ID不能为空")
    private Long examQuestionId;
    
    private String userAnswer;
    
    private Integer answerDurationSeconds;
}
