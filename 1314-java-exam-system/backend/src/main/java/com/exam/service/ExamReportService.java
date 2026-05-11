package com.exam.service;

import com.exam.dto.ExamReportDTO;
import com.exam.entity.*;
import com.exam.enums.ExamStatus;
import com.exam.repository.*;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.*;
import java.util.stream.Collectors;

@Slf4j
@Service
@RequiredArgsConstructor
@Transactional(readOnly = true)
public class ExamReportService {

    private final ExamRecordRepository examRecordRepository;
    private final ExamQuestionRepository examQuestionRepository;
    private final AnswerRecordRepository answerRecordRepository;
    private final ExamExceptionRepository examExceptionRepository;

    public ExamReportDTO generateReport(Long examRecordId) {
        ExamRecord examRecord = examRecordRepository.findById(examRecordId)
                .orElseThrow(() -> new RuntimeException("考试记录不存在: " + examRecordId));

        if (examRecord.getStatus() == ExamStatus.IN_PROGRESS || 
            examRecord.getStatus() == ExamStatus.PAUSED ||
            examRecord.getStatus() == ExamStatus.NOT_STARTED) {
            throw new RuntimeException("考试未结束，无法生成报告");
        }

        ExamPaper examPaper = examRecord.getExamPaper();
        List<ExamQuestion> examQuestions = examQuestionRepository
                .findByExamRecordIdOrderByDisplayOrderAsc(examRecordId);
        
        Map<Long, AnswerRecord> answerRecords = answerRecordRepository.findByExamRecordId(examRecordId)
                .stream()
                .collect(Collectors.toMap(ar -> ar.getExamQuestion().getId(), ar -> ar));

        List<ExamException> exceptions = examExceptionRepository
                .findByExamRecordIdOrderByTimestampAsc(examRecordId);

        int answeredQuestions = 0;
        int correctQuestions = 0;
        int incorrectQuestions = 0;
        
        for (ExamQuestion eq : examQuestions) {
            AnswerRecord ar = answerRecords.get(eq.getId());
            if (ar != null && ar.getUserAnswer() != null && !ar.getUserAnswer().trim().isEmpty()) {
                answeredQuestions++;
                if (ar.getCorrect() != null && ar.getCorrect()) {
                    correctQuestions++;
                } else if (ar.getCorrect() != null) {
                    incorrectQuestions++;
                }
            }
        }

        Map<String, List<ExamQuestion>> questionsByCategory = examQuestions.stream()
                .collect(Collectors.groupingBy(ExamQuestion::getCategoryName));
        
        List<ExamReportDTO.CategoryScoreDTO> categoryScores = new ArrayList<>();
        
        for (Map.Entry<String, List<ExamQuestion>> entry : questionsByCategory.entrySet()) {
            String categoryName = entry.getKey();
            List<ExamQuestion> categoryQuestions = entry.getValue();
            
            int categoryTotalScore = 0;
            int categoryObtainedScore = 0;
            int categoryCorrect = 0;
            int categoryIncorrect = 0;
            int categoryUnanswered = 0;
            
            for (ExamQuestion eq : categoryQuestions) {
                categoryTotalScore += eq.getScore();
                AnswerRecord ar = answerRecords.get(eq.getId());
                
                if (ar != null && ar.getUserAnswer() != null && !ar.getUserAnswer().trim().isEmpty()) {
                    categoryObtainedScore += (ar.getScore() != null ? ar.getScore() : 0);
                    if (ar.getCorrect() != null && ar.getCorrect()) {
                        categoryCorrect++;
                    } else if (ar.getCorrect() != null) {
                        categoryIncorrect++;
                    }
                } else {
                    categoryUnanswered++;
                }
            }
            
            double scoreRate = categoryTotalScore > 0 ? 
                    (double) categoryObtainedScore / categoryTotalScore * 100 : 0;
            
            categoryScores.add(ExamReportDTO.CategoryScoreDTO.builder()
                    .categoryName(categoryName)
                    .totalQuestions(categoryQuestions.size())
                    .totalScore(categoryTotalScore)
                    .obtainedScore(categoryObtainedScore)
                    .scoreRate(scoreRate)
                    .correctCount(categoryCorrect)
                    .incorrectCount(categoryIncorrect)
                    .unansweredCount(categoryUnanswered)
                    .build());
        }

        categoryScores.sort(Comparator.comparing(ExamReportDTO.CategoryScoreDTO::getScoreRate));

        List<ExamReportDTO.ExceptionRecordDTO> exceptionRecords = exceptions.stream()
                .map(e -> ExamReportDTO.ExceptionRecordDTO.builder()
                        .type(e.getType())
                        .description(e.getDescription())
                        .timestamp(e.getTimestamp())
                        .durationSeconds(e.getSwitchOutDurationSeconds())
                        .build())
                .collect(Collectors.toList());

        boolean passed = examRecord.getObtainedScore() != null && 
                         examRecord.getObtainedScore() >= examPaper.getPassScore();
        
        double scorePercentage = examPaper.getTotalScore() > 0 ?
                (examRecord.getObtainedScore() != null ? 
                    (double) examRecord.getObtainedScore() / examPaper.getTotalScore() * 100 : 0) : 0;

        return ExamReportDTO.builder()
                .examRecordId(examRecord.getId())
                .examPaperId(examPaper.getId())
                .examPaperTitle(examPaper.getTitle())
                .totalScore(examPaper.getTotalScore())
                .passScore(examPaper.getPassScore())
                .obtainedScore(examRecord.getObtainedScore())
                .passed(passed)
                .scorePercentage(scorePercentage)
                .totalQuestions(examQuestions.size())
                .answeredQuestions(answeredQuestions)
                .correctQuestions(correctQuestions)
                .incorrectQuestions(incorrectQuestions)
                .durationMinutes(examPaper.getDurationMinutes())
                .extendedMinutes(examRecord.getExtendedMinutes())
                .actualDurationSeconds(examRecord.getActualDurationSeconds())
                .status(examRecord.getStatus())
                .startTime(examRecord.getStartTime())
                .endTime(examRecord.getEndTime())
                .switchCount(examRecord.getSwitchCount())
                .maxSwitchDurationSeconds(examRecord.getMaxSwitchDurationSeconds())
                .categoryScores(categoryScores)
                .exceptions(exceptionRecords)
                .hasAbnormalities(!exceptionRecords.isEmpty())
                .build();
    }

    public List<ExamRecord> getUserExamRecords(Long userId) {
        return examRecordRepository.findByUserIdOrderByCreatedAtDesc(userId);
    }

    public List<ExamRecord> getExamPaperRecords(Long examPaperId) {
        return examRecordRepository.findByExamPaperIdOrderByCreatedAtDesc(examPaperId);
    }
}
