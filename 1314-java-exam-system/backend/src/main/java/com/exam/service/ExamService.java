package com.exam.service;

import com.exam.dto.*;
import com.exam.entity.*;
import com.exam.enums.ExamStatus;
import com.exam.enums.QuestionType;
import com.exam.repository.*;
import com.exam.util.ShuffleUtil;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Duration;
import java.time.LocalDateTime;
import java.util.*;
import java.util.stream.Collectors;

@Slf4j
@Service
@RequiredArgsConstructor
@Transactional
public class ExamService {

    private final ExamRecordRepository examRecordRepository;
    private final ExamQuestionRepository examQuestionRepository;
    private final AnswerRecordRepository answerRecordRepository;
    private final ExamExceptionRepository examExceptionRepository;
    private final ExamPaperRepository examPaperRepository;
    private final UserRepository userRepository;
    private final ExamPaperService examPaperService;
    private final GradingService gradingService;

    public ExamDetailDTO startExam(Long examPaperId, Long userId) {
        ExamPaper examPaper = examPaperRepository.findById(examPaperId)
                .orElseThrow(() -> new RuntimeException("试卷不存在: " + examPaperId));

        User user = userRepository.findById(userId)
                .orElseThrow(() -> new RuntimeException("用户不存在: " + userId));

        Optional<ExamRecord> existingRecord = examRecordRepository
                .findByUserIdAndExamPaperIdAndStatusIn(
                        userId,
                        examPaperId,
                        Arrays.asList(ExamStatus.IN_PROGRESS, ExamStatus.PAUSED));

        if (existingRecord.isPresent()) {
            return resumeExam(existingRecord.get().getId());
        }

        if (!examPaper.isCanRetake()) {
            long completedCount = examRecordRepository.countByUserIdAndExamPaperIdAndStatusIn(
                    userId, 
                    examPaperId,
                    Arrays.asList(ExamStatus.SUBMITTED, ExamStatus.TIMEOUT_SUBMITTED, ExamStatus.ABNORMAL_SUBMITTED));
            if (completedCount > 0) {
                throw new RuntimeException("该试卷不允许重考");
            }
        } else if (examPaper.getMaxRetakeCount() != null) {
            long completedCount = examRecordRepository.countByUserIdAndExamPaperIdAndStatusIn(
                    userId, 
                    examPaperId,
                    Arrays.asList(ExamStatus.SUBMITTED, ExamStatus.TIMEOUT_SUBMITTED, ExamStatus.ABNORMAL_SUBMITTED));
            if (completedCount >= examPaper.getMaxRetakeCount()) {
                throw new RuntimeException("已达到最大重考次数: " + examPaper.getMaxRetakeCount());
            }
        }

        int attemptNumber = (int) examRecordRepository.countByUserIdAndExamPaperIdAndStatusIn(
                userId, 
                examPaperId,
                Arrays.asList(ExamStatus.SUBMITTED, ExamStatus.TIMEOUT_SUBMITTED, ExamStatus.ABNORMAL_SUBMITTED)) + 1;

        ExamRecord examRecord = new ExamRecord();
        examRecord.setExamPaper(examPaper);
        examRecord.setUser(user);
        examRecord.setAttemptNumber(attemptNumber);
        examRecord.setStatus(ExamStatus.IN_PROGRESS);
        examRecord.setStartTime(LocalDateTime.now());
        examRecord.setTotalScore(examPaper.getTotalScore());
        examRecord = examRecordRepository.save(examRecord);

        List<Question> selectedQuestions = examPaperService.selectQuestionsForExam(examPaper);
        List<Question> shuffledQuestions = ShuffleUtil.shuffleList(selectedQuestions);

        Map<Long, Integer> questionScoreMap = buildQuestionScoreMap(examPaper);

        int displayOrder = 0;
        for (Question question : shuffledQuestions) {
            ExamQuestion examQuestion = new ExamQuestion();
            examQuestion.setExamRecord(examRecord);
            examQuestion.setQuestion(question);
            examQuestion.setDisplayOrder(displayOrder++);
            examQuestion.setScore(questionScoreMap.getOrDefault(question.getId(), question.getDefaultScore()));
            examQuestion.setQuestionType(question.getType());
            examQuestion.setDifficulty(question.getDifficulty());
            examQuestion.setCategoryName(question.getCategory().getName());

            if (question.getType() != QuestionType.FILL_BLANK && question.getOptions() != null && !question.getOptions().isEmpty()) {
                ShuffleUtil.ShuffledOptions shuffledOptions = ShuffleUtil.shuffleOptions(question);
                examQuestion.setShuffledOptions(ShuffleUtil.serializeShuffledOptions(shuffledOptions));
            }

            examQuestionRepository.save(examQuestion);
        }

        return buildExamDetailDTO(examRecord);
    }

    public ExamDetailDTO resumeExam(Long examRecordId) {
        ExamRecord examRecord = examRecordRepository.findById(examRecordId)
                .orElseThrow(() -> new RuntimeException("考试记录不存在: " + examRecordId));

        if (examRecord.getStatus() != ExamStatus.IN_PROGRESS && examRecord.getStatus() != ExamStatus.PAUSED) {
            throw new RuntimeException("该考试已结束，无法继续");
        }

        int remainingSeconds = calculateRemainingSeconds(examRecord);
        if (remainingSeconds <= 0) {
            autoSubmitExam(examRecordId);
            return buildExamDetailDTO(examRecord);
        }

        if (examRecord.getStatus() == ExamStatus.PAUSED) {
            examRecord.setStatus(ExamStatus.IN_PROGRESS);
            examRecord = examRecordRepository.save(examRecord);
        }

        return buildExamDetailDTO(examRecord);
    }

    public void submitAnswer(AnswerSubmitDTO dto) {
        ExamRecord examRecord = examRecordRepository.findById(dto.getExamRecordId())
                .orElseThrow(() -> new RuntimeException("考试记录不存在: " + dto.getExamRecordId()));

        if (examRecord.getStatus() != ExamStatus.IN_PROGRESS) {
            throw new RuntimeException("考试未在进行中");
        }

        int remainingSeconds = calculateRemainingSeconds(examRecord);
        if (remainingSeconds <= 0) {
            autoSubmitExam(dto.getExamRecordId());
            return;
        }

        ExamQuestion examQuestion = examQuestionRepository.findById(dto.getExamQuestionId())
                .orElseThrow(() -> new RuntimeException("考试题目不存在: " + dto.getExamQuestionId()));

        Optional<AnswerRecord> existingAnswer = answerRecordRepository
                .findByExamRecordIdAndExamQuestionId(dto.getExamRecordId(), dto.getExamQuestionId());

        AnswerRecord answerRecord;
        if (existingAnswer.isPresent()) {
            answerRecord = existingAnswer.get();
        } else {
            answerRecord = new AnswerRecord();
            answerRecord.setExamRecord(examRecord);
            answerRecord.setExamQuestion(examQuestion);
            answerRecord.setAnswerStartTime(LocalDateTime.now());
        }

        String processedAnswer = processAnswer(dto.getUserAnswer(), examQuestion);
        answerRecord.setUserAnswer(processedAnswer);
        answerRecord.setAnswerEndTime(LocalDateTime.now());

        if (dto.getAnswerDurationSeconds() != null) {
            answerRecord.setAnswerDurationSeconds(dto.getAnswerDurationSeconds());
        } else if (answerRecord.getAnswerStartTime() != null) {
            Duration duration = Duration.between(answerRecord.getAnswerStartTime(), answerRecord.getAnswerEndTime());
            answerRecord.setAnswerDurationSeconds((int) duration.getSeconds());
        }

        Question question = examQuestion.getQuestion();
        if (question.getAnswerTimeLimit() != null && answerRecord.getAnswerDurationSeconds() != null) {
            if (answerRecord.getAnswerDurationSeconds() > question.getAnswerTimeLimit() * 60) {
                answerRecord.setTimeLimitExceeded(true);
                recordException(examRecord, "SINGLE_QUESTION_TIMEOUT", 
                    "题目答题超时，限制" + question.getAnswerTimeLimit() + "分钟，实际用时" + (answerRecord.getAnswerDurationSeconds() / 60) + "分钟", 
                    null);
            }
        }

        answerRecordRepository.save(answerRecord);
    }

    public ExamDetailDTO submitExam(Long examRecordId) {
        ExamRecord examRecord = examRecordRepository.findById(examRecordId)
                .orElseThrow(() -> new RuntimeException("考试记录不存在: " + examRecordId));

        if (examRecord.getStatus() != ExamStatus.IN_PROGRESS) {
            throw new RuntimeException("考试未在进行中");
        }

        return finalizeExam(examRecord, ExamStatus.SUBMITTED);
    }

    public ExamDetailDTO autoSubmitExam(Long examRecordId) {
        ExamRecord examRecord = examRecordRepository.findById(examRecordId)
                .orElseThrow(() -> new RuntimeException("考试记录不存在: " + examRecordId));

        return finalizeExam(examRecord, ExamStatus.TIMEOUT_SUBMITTED);
    }

    private ExamDetailDTO finalizeExam(ExamRecord examRecord, ExamStatus status) {
        examRecord.setStatus(status);
        examRecord.setEndTime(LocalDateTime.now());

        if (examRecord.getStartTime() != null) {
            Duration duration = Duration.between(examRecord.getStartTime(), examRecord.getEndTime());
            examRecord.setActualDurationSeconds((int) duration.getSeconds());
        }

        List<ExamQuestion> examQuestions = examQuestionRepository
                .findByExamRecordIdOrderByDisplayOrderAsc(examRecord.getId());

        Map<Long, AnswerRecord> answerRecordMap = answerRecordRepository.findByExamRecordId(examRecord.getId())
                .stream()
                .collect(Collectors.toMap(ar -> ar.getExamQuestion().getId(), ar -> ar));

        int totalObtainedScore = 0;
        for (ExamQuestion examQuestion : examQuestions) {
            AnswerRecord answerRecord = answerRecordMap.get(examQuestion.getId());
            
            if (answerRecord != null && answerRecord.getUserAnswer() != null && !answerRecord.getUserAnswer().trim().isEmpty()) {
                GradingService.GradingResult result = gradingService.gradeQuestion(
                        examQuestion.getQuestion(),
                        answerRecord.getUserAnswer(),
                        examQuestion.getScore(),
                        examQuestion.getShuffledOptions());

                answerRecord.setCorrect(result.isCorrect());
                answerRecord.setScore(result.getScore());
                answerRecordRepository.save(answerRecord);

                totalObtainedScore += result.getScore();
            }
        }

        examRecord.setObtainedScore(totalObtainedScore);
        examRecord = examRecordRepository.save(examRecord);

        return buildExamDetailDTO(examRecord);
    }

    public void reportSwitchOut(SwitchOutReportDTO dto) {
        ExamRecord examRecord = examRecordRepository.findById(dto.getExamRecordId())
                .orElseThrow(() -> new RuntimeException("考试记录不存在: " + dto.getExamRecordId()));

        if (examRecord.getStatus() != ExamStatus.IN_PROGRESS) {
            return;
        }

        examRecord.setSwitchCount(examRecord.getSwitchCount() + 1);
        
        if (dto.getDurationSeconds() != null) {
            if (examRecord.getMaxSwitchDurationSeconds() == null || 
                dto.getDurationSeconds() > examRecord.getMaxSwitchDurationSeconds()) {
                examRecord.setMaxSwitchDurationSeconds(dto.getDurationSeconds());
            }
        }

        recordException(examRecord, "PAGE_SWITCH_OUT", 
            "考生切出考试页面" + (dto.getDurationSeconds() != null ? "，时长" + dto.getDurationSeconds() + "秒" : ""),
            dto.getDurationSeconds());

        examRecordRepository.save(examRecord);

        ExamPaper examPaper = examRecord.getExamPaper();
        boolean shouldAutoSubmit = false;

        if (examRecord.getSwitchCount() > examPaper.getSwitchCountLimit()) {
            shouldAutoSubmit = true;
            recordException(examRecord, "SWITCH_COUNT_EXCEEDED", 
                "切出次数超过限制: " + examPaper.getSwitchCountLimit() + "次", null);
        }

        if (dto.getDurationSeconds() != null && dto.getDurationSeconds() > examPaper.getSwitchDurationLimitSeconds()) {
            shouldAutoSubmit = true;
            recordException(examRecord, "SWITCH_DURATION_EXCEEDED", 
                "单次切出时长超过限制: " + examPaper.getSwitchDurationLimitSeconds() + "秒", 
                dto.getDurationSeconds());
        }

        if (shouldAutoSubmit) {
            finalizeExam(examRecord, ExamStatus.ABNORMAL_SUBMITTED);
        }
    }

    public void extendExamTime(Long examRecordId, int additionalMinutes) {
        ExamRecord examRecord = examRecordRepository.findById(examRecordId)
                .orElseThrow(() -> new RuntimeException("考试记录不存在: " + examRecordId));

        if (examRecord.getStatus() != ExamStatus.IN_PROGRESS && examRecord.getStatus() != ExamStatus.PAUSED) {
            throw new RuntimeException("该考试已结束，无法延长时间");
        }

        examRecord.setExtendedMinutes(examRecord.getExtendedMinutes() + additionalMinutes);
        examRecordRepository.save(examRecord);

        log.info("考试 {} 延长时间 {} 分钟", examRecordId, additionalMinutes);
    }

    private void recordException(ExamRecord examRecord, String type, String description, Integer durationSeconds) {
        ExamException exception = new ExamException();
        exception.setExamRecord(examRecord);
        exception.setType(type);
        exception.setDescription(description);
        exception.setSwitchOutDurationSeconds(durationSeconds);
        exception.setTimestamp(LocalDateTime.now());
        examExceptionRepository.save(exception);
    }

    private String processAnswer(String userAnswer, ExamQuestion examQuestion) {
        if (userAnswer == null || userAnswer.trim().isEmpty()) {
            return userAnswer;
        }

        QuestionType type = examQuestion.getQuestionType();
        if (type == QuestionType.FILL_BLANK) {
            return userAnswer.trim();
        }

        if (examQuestion.getShuffledOptions() != null && !examQuestion.getShuffledOptions().isEmpty()) {
            ShuffleUtil.ShuffleStorageData shuffleData = ShuffleUtil.deserializeShuffledOptions(examQuestion.getShuffledOptions());
            return ShuffleUtil.convertUserAnswerToOriginal(userAnswer.trim(), shuffleData);
        }

        return userAnswer.trim();
    }

    private Map<Long, Integer> buildQuestionScoreMap(ExamPaper examPaper) {
        Map<Long, Integer> scoreMap = new HashMap<>();
        
        for (ExamRule rule : examPaper.getRules()) {
            List<Question> questions = examPaperService.selectQuestionsForRule(rule);
            for (Question question : questions) {
                scoreMap.put(question.getId(), rule.getScorePerQuestion());
            }
        }
        
        return scoreMap;
    }

    private int calculateRemainingSeconds(ExamRecord examRecord) {
        if (examRecord.getStartTime() == null) {
            return 0;
        }

        int totalDurationSeconds = (examRecord.getExamPaper().getDurationMinutes() + examRecord.getExtendedMinutes()) * 60;
        int elapsedSeconds = (int) Duration.between(examRecord.getStartTime(), LocalDateTime.now()).getSeconds();

        return Math.max(0, totalDurationSeconds - elapsedSeconds);
    }

    private ExamDetailDTO buildExamDetailDTO(ExamRecord examRecord) {
        ExamPaper examPaper = examRecord.getExamPaper();
        List<ExamQuestion> examQuestions = examQuestionRepository
                .findByExamRecordIdOrderByDisplayOrderAsc(examRecord.getId());

        final Map<Long, AnswerRecord> answerRecordMap;
        if (examRecord.getStatus() == ExamStatus.SUBMITTED ||
            examRecord.getStatus() == ExamStatus.TIMEOUT_SUBMITTED ||
            examRecord.getStatus() == ExamStatus.ABNORMAL_SUBMITTED) {
            answerRecordMap = answerRecordRepository.findByExamRecordId(examRecord.getId())
                    .stream()
                    .collect(Collectors.toMap(ar -> ar.getExamQuestion().getId(), ar -> ar));
        } else {
            answerRecordMap = new HashMap<>();
        }

        List<ExamQuestionDTO> questionDTOs = examQuestions.stream()
                .map(eq -> buildExamQuestionDTO(eq, answerRecordMap.get(eq.getId())))
                .collect(Collectors.toList());

        return ExamDetailDTO.builder()
                .examRecordId(examRecord.getId())
                .examPaperId(examPaper.getId())
                .examPaperTitle(examPaper.getTitle())
                .totalScore(examPaper.getTotalScore())
                .passScore(examPaper.getPassScore())
                .durationMinutes(examPaper.getDurationMinutes())
                .extendedMinutes(examRecord.getExtendedMinutes())
                .status(examRecord.getStatus())
                .startTime(examRecord.getStartTime())
                .endTime(examRecord.getEndTime())
                .remainingSeconds(calculateRemainingSeconds(examRecord))
                .obtainedScore(examRecord.getObtainedScore())
                .questions(questionDTOs)
                .build();
    }

    private ExamQuestionDTO buildExamQuestionDTO(ExamQuestion examQuestion, AnswerRecord answerRecord) {
        Question question = examQuestion.getQuestion();
        
        List<ExamQuestionDTO.OptionDTO> options = new ArrayList<>();
        if (question.getType() != QuestionType.FILL_BLANK && question.getOptions() != null) {
            if (examQuestion.getShuffledOptions() != null && !examQuestion.getShuffledOptions().isEmpty()) {
                ShuffleUtil.ShuffleStorageData shuffleData = ShuffleUtil.deserializeShuffledOptions(examQuestion.getShuffledOptions());
                
                Map<String, QuestionOption> optionMap = question.getOptions().stream()
                        .collect(Collectors.toMap(QuestionOption::getOptionKey, o -> o));
                
                for (Map.Entry<String, String> entry : shuffleData.getShuffledToOriginalKeyMap().entrySet()) {
                    String shuffledKey = entry.getKey();
                    String originalKey = entry.getValue();
                    QuestionOption option = optionMap.get(originalKey);
                    if (option != null) {
                        options.add(ExamQuestionDTO.OptionDTO.builder()
                                .optionKey(shuffledKey)
                                .content(option.getContent())
                                .build());
                    }
                }
            } else {
                options = question.getOptions().stream()
                        .sorted(Comparator.comparingInt(QuestionOption::getSortOrder))
                        .map(o -> ExamQuestionDTO.OptionDTO.builder()
                                .optionKey(o.getOptionKey())
                                .content(o.getContent())
                                .build())
                        .collect(Collectors.toList());
            }
        }

        ExamQuestionDTO.ExamQuestionDTOBuilder builder = ExamQuestionDTO.builder()
                .examQuestionId(examQuestion.getId())
                .questionType(examQuestion.getQuestionType())
                .difficulty(examQuestion.getDifficulty())
                .categoryName(examQuestion.getCategoryName())
                .content(question.getContent())
                .score(examQuestion.getScore())
                .answerTimeLimit(question.getAnswerTimeLimit())
                .options(options);

        if (answerRecord != null) {
            String displayAnswer = answerRecord.getUserAnswer();
            if (question.getType() != QuestionType.FILL_BLANK && 
                examQuestion.getShuffledOptions() != null && !examQuestion.getShuffledOptions().isEmpty()) {
                ShuffleUtil.ShuffleStorageData shuffleData = ShuffleUtil.deserializeShuffledOptions(examQuestion.getShuffledOptions());
                displayAnswer = ShuffleUtil.convertOriginalAnswerToShuffled(answerRecord.getUserAnswer(), shuffleData);
            }
            
            builder.userAnswer(displayAnswer)
                   .correct(answerRecord.getCorrect())
                   .obtainedScore(answerRecord.getScore());
        }

        return builder.build();
    }

    public ExamDetailDTO getExamDetail(Long examRecordId) {
        ExamRecord examRecord = examRecordRepository.findById(examRecordId)
                .orElseThrow(() -> new RuntimeException("考试记录不存在: " + examRecordId));
        return buildExamDetailDTO(examRecord);
    }
}
