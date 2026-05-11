package com.exam.dto;

import com.exam.enums.DifficultyLevel;
import com.exam.enums.QuestionType;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

import java.util.List;

@Data
public class QuestionDTO {
    private Long id;
    
    @NotNull(message = "题目类型不能为空")
    private QuestionType type;
    
    @NotNull(message = "分类ID不能为空")
    private Long categoryId;
    
    @NotNull(message = "难度等级不能为空")
    private DifficultyLevel difficulty;
    
    @NotBlank(message = "题目内容不能为空")
    private String content;
    
    private int defaultScore = 10;
    
    private Integer answerTimeLimit;
    
    private List<QuestionOptionDTO> options;
    
    private List<String> answers;
    
    private boolean ignoreCase = true;
    
    private String explanation;
}
