package com.recruitment.server.controller;

import com.recruitment.common.response.ApiResponse;
import com.recruitment.common.response.StatisticsResponse;
import com.recruitment.server.service.StatisticsService;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/statistics")
public class StatisticsController {
    private final StatisticsService statisticsService;

    public StatisticsController(StatisticsService statisticsService) {
        this.statisticsService = statisticsService;
    }

    @GetMapping
    public ApiResponse<StatisticsResponse> getStatistics() {
        return statisticsService.getStatistics();
    }
}
