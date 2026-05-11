package com.exam.controller;

import com.exam.dto.*;
import com.exam.entity.ExamRecord;
import com.exam.service.ExamReportService;
import com.exam.service.ExamService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/exams")
@RequiredArgsConstructor
@Tag(name = "考试管理", description = "考试流程相关接口")
public class ExamController {

    private final ExamService examService;
    private final ExamReportService examReportService;

    @PostMapping("/start")
    @Operation(summary = "开始考试", description = "开始考试，会自动抽题并打乱顺序和选项")
    public ApiResponse<ExamDetailDTO> startExam(@Valid @RequestBody StartExamDTO dto) {
        return ApiResponse.success("考试开始", examService.startExam(dto.getExamPaperId(), dto.getUserId()));
    }

    @PostMapping("/{examRecordId}/resume")
    @Operation(summary = "继续考试", description = "继续未完成的考试，支持断点续考")
    public ApiResponse<ExamDetailDTO> resumeExam(@PathVariable Long examRecordId) {
        return ApiResponse.success("继续考试", examService.resumeExam(examRecordId));
    }

    @GetMapping("/{examRecordId}")
    @Operation(summary = "获取考试详情", description = "获取考试的详细信息，包括题目列表")
    public ApiResponse<ExamDetailDTO> getExamDetail(@PathVariable Long examRecordId) {
        return ApiResponse.success(examService.getExamDetail(examRecordId));
    }

    @PostMapping("/answer")
    @Operation(summary = "提交答案", description = "提交单题答案")
    public ApiResponse<Void> submitAnswer(@Valid @RequestBody AnswerSubmitDTO dto) {
        examService.submitAnswer(dto);
        return ApiResponse.success("答案已保存", null);
    }

    @PostMapping("/{examRecordId}/submit")
    @Operation(summary = "提交试卷", description = "考生手动提交试卷")
    public ApiResponse<ExamDetailDTO> submitExam(@PathVariable Long examRecordId) {
        return ApiResponse.success("提交成功", examService.submitExam(examRecordId));
    }

    @PostMapping("/{examRecordId}/auto-submit")
    @Operation(summary = "自动交卷", description = "系统自动交卷（超时等情况）")
    public ApiResponse<ExamDetailDTO> autoSubmitExam(@PathVariable Long examRecordId) {
        return ApiResponse.success("自动交卷成功", examService.autoSubmitExam(examRecordId));
    }

    @PostMapping("/switch-out")
    @Operation(summary = "报告切出", description = "考生切出考试页面时报告")
    public ApiResponse<Void> reportSwitchOut(@Valid @RequestBody SwitchOutReportDTO dto) {
        examService.reportSwitchOut(dto);
        return ApiResponse.success("已记录", null);
    }

    @PostMapping("/{examRecordId}/extend")
    @Operation(summary = "延长考试时间", description = "管理员为特定考生延长考试时间")
    public ApiResponse<Void> extendExamTime(
            @PathVariable Long examRecordId,
            @RequestParam int additionalMinutes) {
        examService.extendExamTime(examRecordId, additionalMinutes);
        return ApiResponse.success("已延长" + additionalMinutes + "分钟", null);
    }

    @GetMapping("/{examRecordId}/report")
    @Operation(summary = "获取考试报告", description = "获取考试详细报告，包括得分、知识点分析、异常记录")
    public ApiResponse<ExamReportDTO> getExamReport(@PathVariable Long examRecordId) {
        return ApiResponse.success(examReportService.generateReport(examRecordId));
    }

    @GetMapping("/user/{userId}/records")
    @Operation(summary = "获取用户考试记录", description = "获取指定用户的所有考试记录")
    public ApiResponse<List<ExamRecord>> getUserExamRecords(@PathVariable Long userId) {
        return ApiResponse.success(examReportService.getUserExamRecords(userId));
    }

    @GetMapping("/paper/{examPaperId}/records")
    @Operation(summary = "获取试卷考试记录", description = "获取指定试卷的所有考试记录")
    public ApiResponse<List<ExamRecord>> getExamPaperRecords(@PathVariable Long examPaperId) {
        return ApiResponse.success(examReportService.getExamPaperRecords(examPaperId));
    }
}
