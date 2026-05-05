package com.example.server.controller;

import com.example.common.dto.OnboardingStatisticsDTO;
import com.example.common.response.ApiResponse;
import com.example.server.service.StatisticsService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/statistics")
public class StatisticsController {

    @Autowired
    private StatisticsService statisticsService;

    @GetMapping
    public ApiResponse<OnboardingStatisticsDTO> getStatistics() {
        OnboardingStatisticsDTO statistics = statisticsService.getStatistics();
        return ApiResponse.success(statistics);
    }
}
