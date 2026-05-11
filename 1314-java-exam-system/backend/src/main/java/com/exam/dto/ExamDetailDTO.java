package com.exam.dto;

import com.exam.enums.ExamStatus;
import lombok.Data;
import lombok.Builder;
import lombok.AllArgsConstructor;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;
import java.util.List;

@Data
@Builder
@AllArgsConstructor
@NoArgsConstructor
public class ExamDetailDTO {
    private Long examRecordId;
    private Long examPaperId;
    private String examPaperTitle;
    private int totalScore;
    private int passScore;
    private int durationMinutes;
    private int extendedMinutes;
    private ExamStatus status;
    private LocalDateTime startTime;
    private LocalDateTime endTime;
    private Integer remainingSeconds;
    private Integer obtainedScore;
    private List<ExamQuestionDTO> questions;
}
