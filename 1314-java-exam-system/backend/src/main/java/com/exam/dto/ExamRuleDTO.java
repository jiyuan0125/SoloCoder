package com.exam.dto;

import com.exam.enums.DifficultyLevel;
import com.exam.enums.QuestionType;
import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

@Data
public class ExamRuleDTO {
    private Long id;
    
    @NotNull(message = "分类ID不能为空")
    private Long categoryId;
    
    @NotNull(message = "题目类型不能为空")
    private QuestionType questionType;
    
    @NotNull(message = "题目数量不能为空")
    @Min(value = 1, message = "题目数量必须大于0")
    private Integer questionCount;
    
    @NotNull(message = "每题分数不能为空")
    @Min(value = 1, message = "每题分数必须大于0")
    private Integer scorePerQuestion;
    
    private DifficultyLevel preferredDifficulty;
    
    private Integer minDifficultyLevel;
    
    private Integer maxDifficultyLevel;
    
    private Integer minSpecificDifficultyCount;
    
    private DifficultyLevel specificDifficulty;
}
