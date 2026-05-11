package com.exam.service;

import com.exam.entity.Question;
import com.exam.entity.QuestionAnswer;
import com.exam.entity.QuestionOption;
import com.exam.enums.QuestionType;
import com.exam.util.ShuffleUtil;
import lombok.Data;
import lombok.Builder;
import lombok.AllArgsConstructor;
import lombok.NoArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.*;
import java.util.stream.Collectors;

@Slf4j
@Service
public class GradingService {

    public GradingResult gradeQuestion(Question question, String userAnswer, int maxScore, String shuffledOptionsJson) {
        if (userAnswer == null || userAnswer.trim().isEmpty()) {
            return GradingResult.builder()
                    .correct(false)
                    .score(0)
                    .build();
        }

        String trimmedAnswer = userAnswer.trim();

        switch (question.getType()) {
            case SINGLE_CHOICE:
                return gradeSingleChoice(question, trimmedAnswer, maxScore);
            case MULTIPLE_CHOICE:
                return gradeMultipleChoice(question, trimmedAnswer, maxScore);
            case TRUE_FALSE:
                return gradeTrueFalse(question, trimmedAnswer, maxScore);
            case FILL_BLANK:
                return gradeFillBlank(question, trimmedAnswer, maxScore);
            default:
                return GradingResult.builder()
                        .correct(false)
                        .score(0)
                        .build();
        }
    }

    private GradingResult gradeSingleChoice(Question question, String userAnswer, int maxScore) {
        Set<String> correctAnswers = getCorrectOptionKeys(question);
        
        boolean isCorrect = correctAnswers.contains(userAnswer);
        
        return GradingResult.builder()
                .correct(isCorrect)
                .score(isCorrect ? maxScore : 0)
                .build();
    }

    private GradingResult gradeMultipleChoice(Question question, String userAnswer, int maxScore) {
        Set<String> correctAnswers = getCorrectOptionKeys(question);
        Set<String> userAnswers = Arrays.stream(userAnswer.split(","))
                .map(String::trim)
                .filter(s -> !s.isEmpty())
                .collect(Collectors.toSet());

        if (userAnswers.isEmpty()) {
            return GradingResult.builder()
                    .correct(false)
                    .score(0)
                    .build();
        }

        Set<String> intersection = new HashSet<>(userAnswers);
        intersection.retainAll(correctAnswers);

        Set<String> wrongAnswers = new HashSet<>(userAnswers);
        wrongAnswers.removeAll(correctAnswers);

        boolean allCorrect = intersection.equals(correctAnswers);
        boolean hasWrongAnswers = !wrongAnswers.isEmpty();

        if (hasWrongAnswers) {
            return GradingResult.builder()
                    .correct(false)
                    .score(0)
                    .build();
        }

        if (allCorrect) {
            return GradingResult.builder()
                    .correct(true)
                    .score(maxScore)
                    .build();
        } else {
            return GradingResult.builder()
                    .correct(false)
                    .score(maxScore / 2)
                    .build();
        }
    }

    private GradingResult gradeTrueFalse(Question question, String userAnswer, int maxScore) {
        Set<String> correctAnswers = getCorrectOptionKeys(question);
        
        boolean isCorrect = correctAnswers.contains(userAnswer) || 
                           correctAnswers.stream().anyMatch(ca -> 
                               ca.equalsIgnoreCase(userAnswer));
        
        return GradingResult.builder()
                .correct(isCorrect)
                .score(isCorrect ? maxScore : 0)
                .build();
    }

    private GradingResult gradeFillBlank(Question question, String userAnswer, int maxScore) {
        List<QuestionAnswer> correctAnswers = question.getAnswers();
        
        if (correctAnswers == null || correctAnswers.isEmpty()) {
            return GradingResult.builder()
                    .correct(false)
                    .score(0)
                    .build();
        }

        String normalizedUserAnswer = normalizeFillBlankAnswer(userAnswer);
        
        boolean isCorrect = correctAnswers.stream()
                .anyMatch(ca -> {
                    String normalizedCorrect = normalizeFillBlankAnswer(ca.getAnswer());
                    if (question.isIgnoreCase()) {
                        return normalizedUserAnswer.equalsIgnoreCase(normalizedCorrect);
                    } else {
                        return normalizedUserAnswer.equals(normalizedCorrect);
                    }
                });

        return GradingResult.builder()
                .correct(isCorrect)
                .score(isCorrect ? maxScore : 0)
                .build();
    }

    private Set<String> getCorrectOptionKeys(Question question) {
        if (question.getOptions() == null || question.getOptions().isEmpty()) {
            return new HashSet<>();
        }
        
        return question.getOptions().stream()
                .filter(QuestionOption::isCorrect)
                .map(QuestionOption::getOptionKey)
                .collect(Collectors.toSet());
    }

    private String normalizeFillBlankAnswer(String answer) {
        if (answer == null) {
            return "";
        }
        
        return answer
                .trim()
                .replaceAll("\\s+", " ")
                .replaceAll("[，,。.；;：:、]", "");
    }

    @Data
    @Builder
    @AllArgsConstructor
    @NoArgsConstructor
    public static class GradingResult {
        private boolean correct;
        private int score;
    }
}
