package com.exam.controller;

import com.exam.dto.ApiResponse;
import com.exam.dto.QuestionDTO;
import com.exam.entity.Question;
import com.exam.service.QuestionService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/questions")
@RequiredArgsConstructor
@Tag(name = "题库管理", description = "题目的增删改查接口")
public class QuestionController {

    private final QuestionService questionService;

    @GetMapping
    @Operation(summary = "获取所有题目", description = "获取题库中所有题目")
    public ApiResponse<List<Question>> getAllQuestions() {
        return ApiResponse.success(questionService.getAllQuestions());
    }

    @GetMapping("/{id}")
    @Operation(summary = "获取单个题目", description = "根据ID获取题目详情")
    public ApiResponse<Question> getQuestionById(@PathVariable Long id) {
        return ApiResponse.success(questionService.getQuestionById(id));
    }

    @GetMapping("/category/{categoryId}")
    @Operation(summary = "按分类获取题目", description = "根据分类ID获取该分类下的所有题目")
    public ApiResponse<List<Question>> getQuestionsByCategory(@PathVariable Long categoryId) {
        return ApiResponse.success(questionService.getQuestionsByCategory(categoryId));
    }

    @PostMapping
    @Operation(summary = "创建题目", description = "创建新题目，支持单选、多选、判断、填空四种类型")
    public ApiResponse<Question> createQuestion(@Valid @RequestBody QuestionDTO dto) {
        return ApiResponse.success("创建成功", questionService.createQuestion(dto));
    }

    @PutMapping("/{id}")
    @Operation(summary = "更新题目", description = "更新题目信息")
    public ApiResponse<Question> updateQuestion(
            @PathVariable Long id,
            @Valid @RequestBody QuestionDTO dto) {
        return ApiResponse.success("更新成功", questionService.updateQuestion(id, dto));
    }

    @DeleteMapping("/{id}")
    @Operation(summary = "删除题目", description = "删除题目")
    public ApiResponse<Void> deleteQuestion(@PathVariable Long id) {
        questionService.deleteQuestion(id);
        return ApiResponse.success("删除成功", null);
    }
}
