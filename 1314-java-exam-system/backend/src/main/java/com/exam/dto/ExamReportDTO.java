package com.exam.dto;

import com.exam.enums.ExamStatus;
import lombok.Data;
import lombok.Builder;
import lombok.AllArgsConstructor;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Map;

@Data
@Builder
@AllArgsConstructor
@NoArgsConstructor
public class ExamReportDTO {
    private Long examRecordId;
    private Long examPaperId;
    private String examPaperTitle;
    private int totalScore;
    private int passScore;
    private Integer obtainedScore;
    private boolean passed;
    private double scorePercentage;
    
    private int totalQuestions;
    private int answeredQuestions;
    private int correctQuestions;
    private int incorrectQuestions;
    
    private int durationMinutes;
    private int extendedMinutes;
    private Integer actualDurationSeconds;
    
    private ExamStatus status;
    private LocalDateTime startTime;
    private LocalDateTime endTime;
    
    private int switchCount;
    private Integer maxSwitchDurationSeconds;
    
    private List<CategoryScoreDTO> categoryScores;
    private List<ExceptionRecordDTO> exceptions;
    private boolean hasAbnormalities;
    
    @Data
    @Builder
    @AllArgsConstructor
    @NoArgsConstructor
    public static class CategoryScoreDTO {
        private String categoryName;
        private int totalQuestions;
        private int totalScore;
        private int obtainedScore;
        private double scoreRate;
        private int correctCount;
        private int incorrectCount;
        private int unansweredCount;
    }
    
    @Data
    @Builder
    @AllArgsConstructor
    @NoArgsConstructor
    public static class ExceptionRecordDTO {
        private String type;
        private String description;
        private LocalDateTime timestamp;
        private Integer durationSeconds;
    }
}
