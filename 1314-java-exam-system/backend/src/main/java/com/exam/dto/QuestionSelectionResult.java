package com.exam.dto;

import lombok.Data;
import lombok.Builder;
import lombok.AllArgsConstructor;
import lombok.NoArgsConstructor;

import java.util.List;

@Data
@Builder
@AllArgsConstructor
@NoArgsConstructor
public class QuestionSelectionResult {
    private boolean success;
    private String message;
    private List<SelectedQuestionInfo> selectedQuestions;
    
    @Data
    @Builder
    @AllArgsConstructor
    @NoArgsConstructor
    public static class SelectedQuestionInfo {
        private Long questionId;
        private String categoryName;
        private String questionType;
        private int difficultyLevel;
        private int score;
    }
    
    @Data
    @Builder
    @AllArgsConstructor
    @NoArgsConstructor
    public static class RuleValidationResult {
        private Long categoryId;
        private String categoryName;
        private String questionType;
        private int requestedCount;
        private int availableCount;
        private boolean sufficient;
        private String message;
    }
}
