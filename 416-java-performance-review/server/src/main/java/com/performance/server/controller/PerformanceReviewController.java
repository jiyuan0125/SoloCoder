package com.performance.server.controller;

import com.performance.common.constant.ErrorCode;
import com.performance.common.dto.ApiResponse;
import com.performance.common.dto.PerformanceReviewDTO;
import com.performance.common.request.ConfirmReviewRequest;
import com.performance.common.request.QueryHistoryRequest;
import com.performance.common.request.SubmitManagerReviewRequest;
import com.performance.common.request.SubmitSelfReviewRequest;
import com.performance.server.service.PerformanceReviewService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;
import java.util.Optional;

@RestController
@RequestMapping("/api/reviews")
public class PerformanceReviewController {

    @Autowired
    private PerformanceReviewService performanceReviewService;

    @PostMapping("/self-review")
    public ApiResponse<Void> submitSelfReview(@RequestBody SubmitSelfReviewRequest request) {
        Optional<ErrorCode> error = performanceReviewService.submitSelfReview(request);
        if (error.isPresent()) {
            return ApiResponse.error(error.get());
        }
        return ApiResponse.success();
    }

    @PostMapping("/manager-review")
    public ApiResponse<Void> submitManagerReview(@RequestBody SubmitManagerReviewRequest request) {
        Optional<ErrorCode> error = performanceReviewService.submitManagerReview(request);
        if (error.isPresent()) {
            return ApiResponse.error(error.get());
        }
        return ApiResponse.success();
    }

    @PostMapping("/original-comment")
    public ApiResponse<Void> addOriginalDepartmentComment(
            @RequestParam Long cycleId,
            @RequestParam Long employeeId,
            @RequestParam Long originalDeptManagerId,
            @RequestParam(required = false) String comment) {
        Optional<ErrorCode> error = performanceReviewService.addOriginalDepartmentComment(
                cycleId, employeeId, originalDeptManagerId, comment);
        if (error.isPresent()) {
            return ApiResponse.error(error.get());
        }
        return ApiResponse.success();
    }

    @PostMapping("/confirm")
    public ApiResponse<Void> confirmReview(@RequestBody ConfirmReviewRequest request) {
        Optional<ErrorCode> error = performanceReviewService.confirmReview(request);
        if (error.isPresent()) {
            return ApiResponse.error(error.get());
        }
        return ApiResponse.success();
    }

    @PostMapping("/calculate-rankings")
    public ApiResponse<Void> calculateRankings(
            @RequestParam Long cycleId,
            @RequestParam Long departmentId) {
        performanceReviewService.calculateRankings(cycleId, departmentId);
        return ApiResponse.success();
    }

    @GetMapping("/{id}")
    public ApiResponse<PerformanceReviewDTO> getReviewById(
            @PathVariable Long id,
            @RequestParam(required = false, defaultValue = "false") boolean isEmployeeView) {
        Optional<PerformanceReviewDTO> reviewOpt = performanceReviewService.getReviewById(id, isEmployeeView);
        if (reviewOpt.isPresent()) {
            return ApiResponse.success(reviewOpt.get());
        }
        return ApiResponse.error(ErrorCode.NOT_FOUND);
    }

    @GetMapping("/cycle/{cycleId}/employee/{employeeId}")
    public ApiResponse<PerformanceReviewDTO> getReviewByCycleAndEmployee(
            @PathVariable Long cycleId,
            @PathVariable Long employeeId,
            @RequestParam(required = false, defaultValue = "false") boolean isEmployeeView) {
        Optional<PerformanceReviewDTO> reviewOpt = performanceReviewService.getReviewByCycleAndEmployee(
                cycleId, employeeId, isEmployeeView);
        if (reviewOpt.isPresent()) {
            return ApiResponse.success(reviewOpt.get());
        }
        return ApiResponse.error(ErrorCode.NOT_FOUND);
    }

    @GetMapping("/cycle/{cycleId}")
    public ApiResponse<List<PerformanceReviewDTO>> getReviewsByCycleId(@PathVariable Long cycleId) {
        List<PerformanceReviewDTO> reviews = performanceReviewService.getReviewsByCycleId(cycleId);
        return ApiResponse.success(reviews);
    }

    @GetMapping("/employee/{employeeId}")
    public ApiResponse<List<PerformanceReviewDTO>> getReviewsByEmployeeId(@PathVariable Long employeeId) {
        List<PerformanceReviewDTO> reviews = performanceReviewService.getReviewsByEmployeeId(employeeId);
        return ApiResponse.success(reviews);
    }

    @PostMapping("/history")
    public ApiResponse<List<PerformanceReviewDTO>> queryHistory(@RequestBody QueryHistoryRequest request) {
        List<PerformanceReviewDTO> reviews = performanceReviewService.queryHistory(request);
        return ApiResponse.success(reviews);
    }
}