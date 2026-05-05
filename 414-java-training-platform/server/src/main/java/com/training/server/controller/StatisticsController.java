package com.training.server.controller;

import com.training.common.response.ApiResponse;
import com.training.common.dto.response.CreditSummaryDTO;
import com.training.common.dto.response.InstructorEvaluationSummaryDTO;
import com.training.common.dto.response.StatisticsDTO;
import com.training.server.service.StatisticsService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/statistics")
public class StatisticsController {

    @Autowired
    private StatisticsService statisticsService;

    @GetMapping
    public ApiResponse<StatisticsDTO> getOverallStatistics() {
        StatisticsDTO statistics = statisticsService.getOverallStatistics();
        return ApiResponse.success(statistics);
    }

    @GetMapping("/credits/{employeeId}")
    public ApiResponse<CreditSummaryDTO> getCreditSummary(
            @PathVariable String employeeId,
            @RequestParam(required = false) Integer year) {
        int actualYear = year != null ? year : java.time.Year.now().getValue();
        CreditSummaryDTO summary = statisticsService.getCreditSummary(employeeId, actualYear);
        return ApiResponse.success(summary);
    }

    @GetMapping("/evaluations/{instructorId}")
    public ApiResponse<InstructorEvaluationSummaryDTO> getInstructorEvaluations(@PathVariable String instructorId) {
        InstructorEvaluationSummaryDTO summary = statisticsService.getInstructorEvaluations(instructorId);
        return ApiResponse.success(summary);
    }
}
