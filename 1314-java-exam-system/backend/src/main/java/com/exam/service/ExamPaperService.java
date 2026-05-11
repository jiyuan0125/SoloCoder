package com.exam.service;

import com.exam.dto.ExamPaperDTO;
import com.exam.dto.ExamRuleDTO;
import com.exam.dto.QuestionSelectionResult;
import com.exam.entity.ExamPaper;
import com.exam.entity.ExamRule;
import com.exam.entity.KnowledgeCategory;
import com.exam.entity.Question;
import com.exam.enums.DifficultyLevel;
import com.exam.repository.ExamPaperRepository;
import com.exam.repository.ExamRuleRepository;
import com.exam.repository.KnowledgeCategoryRepository;
import com.exam.repository.QuestionRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.*;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
@Transactional
public class ExamPaperService {

    private final ExamPaperRepository examPaperRepository;
    private final ExamRuleRepository examRuleRepository;
    private final KnowledgeCategoryRepository categoryRepository;
    private final QuestionRepository questionRepository;

    public List<ExamPaper> getAllExamPapers() {
        return examPaperRepository.findByEnabledTrueOrderByCreatedAtDesc();
    }

    public ExamPaper getExamPaperById(Long id) {
        return examPaperRepository.findById(id)
                .orElseThrow(() -> new RuntimeException("试卷不存在: " + id));
    }

    public ExamPaper createExamPaper(ExamPaperDTO dto) {
        validateExamPaperRules(dto);
        
        QuestionSelectionResult validationResult = validateQuestionAvailability(dto);
        if (!validationResult.isSuccess()) {
            throw new RuntimeException(validationResult.getMessage());
        }

        ExamPaper examPaper = new ExamPaper();
        examPaper.setTitle(dto.getTitle());
        examPaper.setDescription(dto.getDescription());
        examPaper.setTotalScore(dto.getTotalScore());
        examPaper.setPassScore(dto.getPassScore());
        examPaper.setDurationMinutes(dto.getDurationMinutes());
        examPaper.setCanRetake(dto.isCanRetake());
        examPaper.setMaxRetakeCount(dto.getMaxRetakeCount());
        examPaper.setSwitchCountLimit(dto.getSwitchCountLimit() != null ? dto.getSwitchCountLimit() : 3);
        examPaper.setSwitchDurationLimitSeconds(dto.getSwitchDurationLimitSeconds() != null ? dto.getSwitchDurationLimitSeconds() : 30);
        examPaper.setEnabled(dto.isEnabled());

        if (dto.getRules() != null && !dto.getRules().isEmpty()) {
            List<ExamRule> rules = new ArrayList<>();
            for (ExamRuleDTO ruleDTO : dto.getRules()) {
                ExamRule rule = convertToRuleEntity(ruleDTO, examPaper);
                rules.add(rule);
            }
            examPaper.setRules(rules);
        }

        return examPaperRepository.save(examPaper);
    }

    public ExamPaper updateExamPaper(Long id, ExamPaperDTO dto) {
        ExamPaper examPaper = getExamPaperById(id);
        validateExamPaperRules(dto);

        examPaper.setTitle(dto.getTitle());
        examPaper.setDescription(dto.getDescription());
        examPaper.setTotalScore(dto.getTotalScore());
        examPaper.setPassScore(dto.getPassScore());
        examPaper.setDurationMinutes(dto.getDurationMinutes());
        examPaper.setCanRetake(dto.isCanRetake());
        examPaper.setMaxRetakeCount(dto.getMaxRetakeCount());
        examPaper.setSwitchCountLimit(dto.getSwitchCountLimit() != null ? dto.getSwitchCountLimit() : 3);
        examPaper.setSwitchDurationLimitSeconds(dto.getSwitchDurationLimitSeconds() != null ? dto.getSwitchDurationLimitSeconds() : 30);
        examPaper.setEnabled(dto.isEnabled());

        examPaper.getRules().clear();
        if (dto.getRules() != null && !dto.getRules().isEmpty()) {
            for (ExamRuleDTO ruleDTO : dto.getRules()) {
                ExamRule rule = convertToRuleEntity(ruleDTO, examPaper);
                examPaper.getRules().add(rule);
            }
        }

        return examPaperRepository.save(examPaper);
    }

    public void deleteExamPaper(Long id) {
        ExamPaper examPaper = getExamPaperById(id);
        examPaperRepository.delete(examPaper);
    }

    private void validateExamPaperRules(ExamPaperDTO dto) {
        if (dto.getRules() == null || dto.getRules().isEmpty()) {
            throw new RuntimeException("试卷至少需要一个组卷规则");
        }

        int calculatedTotalScore = 0;
        for (ExamRuleDTO rule : dto.getRules()) {
            calculatedTotalScore += rule.getQuestionCount() * rule.getScorePerQuestion();
        }

        if (calculatedTotalScore != dto.getTotalScore()) {
            throw new RuntimeException("规则计算的总分(" + calculatedTotalScore + ")与试卷总分(" + dto.getTotalScore() + ")不一致");
        }

        if (dto.getPassScore() > dto.getTotalScore()) {
            throw new RuntimeException("及格分不能超过总分");
        }
    }

    public QuestionSelectionResult validateQuestionAvailability(ExamPaperDTO dto) {
        List<QuestionSelectionResult.RuleValidationResult> validationResults = new ArrayList<>();
        boolean allSufficient = true;
        StringBuilder messageBuilder = new StringBuilder();

        for (ExamRuleDTO rule : dto.getRules()) {
            KnowledgeCategory category = categoryRepository.findById(rule.getCategoryId())
                    .orElseThrow(() -> new RuntimeException("分类不存在: " + rule.getCategoryId()));

            List<Question> candidateQuestions = getCandidateQuestions(rule);
            int availableCount = candidateQuestions.size();
            boolean sufficient = availableCount >= rule.getQuestionCount();

            if (!sufficient) {
                allSufficient = false;
                if (messageBuilder.length() > 0) {
                    messageBuilder.append("; ");
                }
                messageBuilder.append(String.format("[%s-%s] 需要%d题，实际只有%d题",
                        category.getName(),
                        rule.getQuestionType(),
                        rule.getQuestionCount(),
                        availableCount));
            }

            validationResults.add(QuestionSelectionResult.RuleValidationResult.builder()
                    .categoryId(rule.getCategoryId())
                    .categoryName(category.getName())
                    .questionType(rule.getQuestionType().name())
                    .requestedCount(rule.getQuestionCount())
                    .availableCount(availableCount)
                    .sufficient(sufficient)
                    .message(sufficient ? "题目充足" : "题目不足")
                    .build());
        }

        return QuestionSelectionResult.builder()
                .success(allSufficient)
                .message(allSufficient ? "所有规则题目充足" : "题目不足: " + messageBuilder.toString())
                .build();
    }

    private List<Question> getCandidateQuestions(ExamRuleDTO rule) {
        if (rule.getMinDifficultyLevel() != null && rule.getMaxDifficultyLevel() != null) {
            List<DifficultyLevel> difficulties = getDifficultyLevelsInRange(
                    rule.getMinDifficultyLevel(), 
                    rule.getMaxDifficultyLevel()
            );
            return questionRepository.findByCategoryIdAndTypeAndDifficultyIn(
                    rule.getCategoryId(),
                    rule.getQuestionType(),
                    difficulties
            );
        } else if (rule.getPreferredDifficulty() != null) {
            return questionRepository.findByCategoryIdAndTypeAndDifficulty(
                    rule.getCategoryId(),
                    rule.getQuestionType(),
                    rule.getPreferredDifficulty()
            );
        } else {
            return questionRepository.findByCategoryIdAndType(
                    rule.getCategoryId(),
                    rule.getQuestionType()
            );
        }
    }

    public List<Question> selectQuestionsForExam(ExamPaper examPaper) {
        List<Question> allSelectedQuestions = new ArrayList<>();

        for (ExamRule rule : examPaper.getRules()) {
            List<Question> ruleQuestions = selectQuestionsForRule(rule);
            allSelectedQuestions.addAll(ruleQuestions);
        }

        return allSelectedQuestions;
    }

    public List<Question> selectQuestionsForRule(ExamRule rule) {
        List<Question> candidates = getCandidateQuestionsForRule(rule);
        
        if (candidates.size() < rule.getQuestionCount()) {
            throw new RuntimeException(String.format("[%s-%s] 题目不足，需要%d题，实际只有%d题",
                    rule.getCategory().getName(),
                    rule.getQuestionType(),
                    rule.getQuestionCount(),
                    candidates.size()));
        }

        List<Question> selected = new ArrayList<>();

        if (rule.getMinSpecificDifficultyCount() != null && rule.getSpecificDifficulty() != null) {
            List<Question> specificDifficultyQuestions = candidates.stream()
                    .filter(q -> q.getDifficulty() == rule.getSpecificDifficulty())
                    .collect(Collectors.toList());

            if (specificDifficultyQuestions.size() < rule.getMinSpecificDifficultyCount()) {
                throw new RuntimeException(String.format("[%s-%s] 特定难度题目不足，需要至少%d题",
                        rule.getCategory().getName(),
                        rule.getQuestionType(),
                        rule.getMinSpecificDifficultyCount()));
            }

            Collections.shuffle(specificDifficultyQuestions);
            selected.addAll(specificDifficultyQuestions.subList(0, rule.getMinSpecificDifficultyCount()));

            List<Question> remaining = new ArrayList<>(candidates);
            remaining.removeAll(specificDifficultyQuestions);
            Collections.shuffle(remaining);

            int remainingNeeded = rule.getQuestionCount() - rule.getMinSpecificDifficultyCount();
            if (remainingNeeded > 0) {
                if (remaining.size() < remainingNeeded) {
                    throw new RuntimeException(String.format("[%s-%s] 剩余题目不足",
                            rule.getCategory().getName(),
                            rule.getQuestionType()));
                }
                selected.addAll(remaining.subList(0, remainingNeeded));
            }
        } else {
            if (rule.getPreferredDifficulty() != null) {
                List<Question> preferredQuestions = candidates.stream()
                        .filter(q -> q.getDifficulty() == rule.getPreferredDifficulty())
                        .collect(Collectors.toList());
                List<Question> otherQuestions = new ArrayList<>(candidates);
                otherQuestions.removeAll(preferredQuestions);

                Collections.shuffle(preferredQuestions);
                Collections.shuffle(otherQuestions);

                selected.addAll(preferredQuestions);
                selected.addAll(otherQuestions);
                
                if (selected.size() > rule.getQuestionCount()) {
                    selected = selected.subList(0, rule.getQuestionCount());
                }
            } else {
                Collections.shuffle(candidates);
                selected.addAll(candidates.subList(0, rule.getQuestionCount()));
            }
        }

        return selected;
    }

    private List<Question> getCandidateQuestionsForRule(ExamRule rule) {
        if (rule.getMinDifficultyLevel() != null && rule.getMaxDifficultyLevel() != null) {
            List<DifficultyLevel> difficulties = getDifficultyLevelsInRange(
                    rule.getMinDifficultyLevel(),
                    rule.getMaxDifficultyLevel()
            );
            return questionRepository.findByCategoryIdAndTypeAndDifficultyIn(
                    rule.getCategory().getId(),
                    rule.getQuestionType(),
                    difficulties
            );
        } else {
            return questionRepository.findByCategoryIdAndType(
                    rule.getCategory().getId(),
                    rule.getQuestionType()
            );
        }
    }

    private List<DifficultyLevel> getDifficultyLevelsInRange(int minLevel, int maxLevel) {
        List<DifficultyLevel> result = new ArrayList<>();
        for (DifficultyLevel level : DifficultyLevel.values()) {
            if (level.getLevel() >= minLevel && level.getLevel() <= maxLevel) {
                result.add(level);
            }
        }
        return result;
    }

    private ExamRule convertToRuleEntity(ExamRuleDTO dto, ExamPaper examPaper) {
        KnowledgeCategory category = categoryRepository.findById(dto.getCategoryId())
                .orElseThrow(() -> new RuntimeException("分类不存在: " + dto.getCategoryId()));

        ExamRule rule = new ExamRule();
        rule.setExamPaper(examPaper);
        rule.setCategory(category);
        rule.setQuestionType(dto.getQuestionType());
        rule.setQuestionCount(dto.getQuestionCount());
        rule.setScorePerQuestion(dto.getScorePerQuestion());
        rule.setPreferredDifficulty(dto.getPreferredDifficulty());
        rule.setMinDifficultyLevel(dto.getMinDifficultyLevel());
        rule.setMaxDifficultyLevel(dto.getMaxDifficultyLevel());
        rule.setMinSpecificDifficultyCount(dto.getMinSpecificDifficultyCount());
        rule.setSpecificDifficulty(dto.getSpecificDifficulty());

        return rule;
    }
}
