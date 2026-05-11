package com.exam.controller;

import com.exam.dto.ApiResponse;
import com.exam.dto.ExamPaperDTO;
import com.exam.dto.QuestionSelectionResult;
import com.exam.entity.ExamPaper;
import com.exam.service.ExamPaperService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/exam-papers")
@RequiredArgsConstructor
@Tag(name = "试卷管理", description = "试卷的增删改查和组卷接口")
public class ExamPaperController {

    private final ExamPaperService examPaperService;

    @GetMapping
    @Operation(summary = "获取所有试卷", description = "获取所有启用的试卷列表")
    public ApiResponse<List<ExamPaper>> getAllExamPapers() {
        return ApiResponse.success(examPaperService.getAllExamPapers());
    }

    @GetMapping("/{id}")
    @Operation(summary = "获取单个试卷", description = "根据ID获取试卷详情，包括组卷规则")
    public ApiResponse<ExamPaper> getExamPaperById(@PathVariable Long id) {
        return ApiResponse.success(examPaperService.getExamPaperById(id));
    }

    @PostMapping
    @Operation(summary = "创建试卷", description = "创建新试卷，会自动验证题库中是否有足够的题目")
    public ApiResponse<ExamPaper> createExamPaper(@Valid @RequestBody ExamPaperDTO dto) {
        return ApiResponse.success("创建成功", examPaperService.createExamPaper(dto));
    }

    @PutMapping("/{id}")
    @Operation(summary = "更新试卷", description = "更新试卷信息和组卷规则")
    public ApiResponse<ExamPaper> updateExamPaper(
            @PathVariable Long id,
            @Valid @RequestBody ExamPaperDTO dto) {
        return ApiResponse.success("更新成功", examPaperService.updateExamPaper(id, dto));
    }

    @DeleteMapping("/{id}")
    @Operation(summary = "删除试卷", description = "删除试卷")
    public ApiResponse<Void> deleteExamPaper(@PathVariable Long id) {
        examPaperService.deleteExamPaper(id);
        return ApiResponse.success("删除成功", null);
    }

    @PostMapping("/validate")
    @Operation(summary = "验证题目是否充足", description = "在创建试卷前验证题库中是否有足够的题目满足组卷规则")
    public ApiResponse<QuestionSelectionResult> validateQuestionAvailability(
            @Valid @RequestBody ExamPaperDTO dto) {
        return ApiResponse.success(examPaperService.validateQuestionAvailability(dto));
    }
}
