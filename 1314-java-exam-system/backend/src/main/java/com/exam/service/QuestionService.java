package com.exam.service;

import com.exam.dto.QuestionDTO;
import com.exam.dto.QuestionOptionDTO;
import com.exam.entity.KnowledgeCategory;
import com.exam.entity.Question;
import com.exam.entity.QuestionAnswer;
import com.exam.entity.QuestionOption;
import com.exam.repository.KnowledgeCategoryRepository;
import com.exam.repository.QuestionRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
@Transactional
public class QuestionService {

    private final QuestionRepository questionRepository;
    private final KnowledgeCategoryRepository categoryRepository;

    public List<Question> getAllQuestions() {
        return questionRepository.findAll();
    }

    public Question getQuestionById(Long id) {
        return questionRepository.findById(id)
                .orElseThrow(() -> new RuntimeException("题目不存在: " + id));
    }

    public Question createQuestion(QuestionDTO dto) {
        KnowledgeCategory category = categoryRepository.findById(dto.getCategoryId())
                .orElseThrow(() -> new RuntimeException("分类不存在: " + dto.getCategoryId()));

        Question question = new Question();
        question.setType(dto.getType());
        question.setCategory(category);
        question.setDifficulty(dto.getDifficulty());
        question.setContent(dto.getContent());
        question.setDefaultScore(dto.getDefaultScore());
        question.setAnswerTimeLimit(dto.getAnswerTimeLimit());
        question.setIgnoreCase(dto.isIgnoreCase());
        question.setExplanation(dto.getExplanation());

        if (dto.getOptions() != null && !dto.getOptions().isEmpty()) {
            List<QuestionOption> options = new ArrayList<>();
            for (int i = 0; i < dto.getOptions().size(); i++) {
                QuestionOptionDTO optDTO = dto.getOptions().get(i);
                QuestionOption option = new QuestionOption();
                option.setQuestion(question);
                option.setOptionKey(optDTO.getOptionKey());
                option.setContent(optDTO.getContent());
                option.setCorrect(optDTO.isCorrect());
                option.setSortOrder(i);
                options.add(option);
            }
            question.setOptions(options);
        }

        if (dto.getAnswers() != null && !dto.getAnswers().isEmpty()) {
            List<QuestionAnswer> answers = new ArrayList<>();
            for (String answerText : dto.getAnswers()) {
                QuestionAnswer answer = new QuestionAnswer();
                answer.setQuestion(question);
                answer.setAnswer(answerText);
                answers.add(answer);
            }
            question.setAnswers(answers);
        }

        return questionRepository.save(question);
    }

    public Question updateQuestion(Long id, QuestionDTO dto) {
        Question question = getQuestionById(id);
        KnowledgeCategory category = categoryRepository.findById(dto.getCategoryId())
                .orElseThrow(() -> new RuntimeException("分类不存在: " + dto.getCategoryId()));

        question.setType(dto.getType());
        question.setCategory(category);
        question.setDifficulty(dto.getDifficulty());
        question.setContent(dto.getContent());
        question.setDefaultScore(dto.getDefaultScore());
        question.setAnswerTimeLimit(dto.getAnswerTimeLimit());
        question.setIgnoreCase(dto.isIgnoreCase());
        question.setExplanation(dto.getExplanation());

        question.getOptions().clear();
        if (dto.getOptions() != null && !dto.getOptions().isEmpty()) {
            for (int i = 0; i < dto.getOptions().size(); i++) {
                QuestionOptionDTO optDTO = dto.getOptions().get(i);
                QuestionOption option = new QuestionOption();
                option.setQuestion(question);
                option.setOptionKey(optDTO.getOptionKey());
                option.setContent(optDTO.getContent());
                option.setCorrect(optDTO.isCorrect());
                option.setSortOrder(i);
                question.getOptions().add(option);
            }
        }

        question.getAnswers().clear();
        if (dto.getAnswers() != null && !dto.getAnswers().isEmpty()) {
            for (String answerText : dto.getAnswers()) {
                QuestionAnswer answer = new QuestionAnswer();
                answer.setQuestion(question);
                answer.setAnswer(answerText);
                question.getAnswers().add(answer);
            }
        }

        return questionRepository.save(question);
    }

    public void deleteQuestion(Long id) {
        Question question = getQuestionById(id);
        questionRepository.delete(question);
    }

    public List<Question> getQuestionsByCategory(Long categoryId) {
        return questionRepository.findByCategoryId(categoryId);
    }
}
