package com.exam.dto;

import com.exam.enums.DifficultyLevel;
import com.exam.enums.QuestionType;
import lombok.Data;
import lombok.Builder;
import lombok.AllArgsConstructor;
import lombok.NoArgsConstructor;

import java.util.List;

@Data
@Builder
@AllArgsConstructor
@NoArgsConstructor
public class ExamQuestionDTO {
    private Long examQuestionId;
    private QuestionType questionType;
    private DifficultyLevel difficulty;
    private String categoryName;
    private String content;
    private int score;
    private Integer answerTimeLimit;
    private List<OptionDTO> options;
    private String userAnswer;
    private Boolean correct;
    private Integer obtainedScore;
    
    @Data
    @Builder
    @AllArgsConstructor
    @NoArgsConstructor
    public static class OptionDTO {
        private String optionKey;
        private String content;
    }
}
